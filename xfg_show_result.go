package main

import (
	"bufio"
	"fmt"
	"sort"
	"strings"

	"github.com/bayashi/colorpalette"
	"github.com/bayashi/xfg/internal/xfgutil"
)

func (x *xfg) setHighlighter() {
	o := x.options
	h := highlighter{}
	if o.ColorPathBase != "" && colorpalette.Exists(o.ColorPathBase) {
		h.pathBaseColor = fmt.Sprintf("\x1b[%sm", colorpalette.GetCode(o.ColorPathBase))
	} else {
		h.pathBaseColor = fmt.Sprintf("\x1b[%sm", colorpalette.GetCode("yellow"))
	}

	if o.ColorPath != "" && colorpalette.Exists(o.ColorPath) {
		h.pathHighlightColor = colorpalette.Get(o.ColorPath)
	} else {
		h.pathHighlightColor = colorpalette.Get("cyan")
	}
	for _, sp := range o.SearchPath {
		h.pathHighlighter = append(h.pathHighlighter, h.pathHighlightColor.Sprintf("%s", sp))
	}

	if o.ColorContent != "" && colorpalette.Exists(o.ColorContent) {
		h.grepHighlightColor = colorpalette.Get(o.ColorContent)
	} else {
		h.grepHighlightColor = colorpalette.Get("red")
	}
	for _, sg := range o.SearchGrep {
		h.grepHighlighter = append(h.grepHighlighter, h.grepHighlightColor.Sprintf("%s", sg))
	}

	x.highlighter = h
}

func (x *xfg) highlightPath(fPath string) string {
	h := x.highlighter

	if len(x.extra.searchPathRe) > 0 {
		for _, re := range x.extra.searchPathRe {
			fPath = re.ReplaceAllString(fPath, h.pathHighlightColor.Sprintf("$1")+h.pathBaseColor)
		}
	}

	if x.options.IgnoreCase {
		for _, spr := range x.extra.searchPathi {
			fPath = spr.ReplaceAllString(fPath, h.pathHighlightColor.Sprintf("$1")+h.pathBaseColor)
		}
	} else {
		for i, sp := range x.options.SearchPath {
			fPath = strings.ReplaceAll(fPath, sp, h.pathHighlighter[i]+h.pathBaseColor)
		}
	}

	return h.pathBaseColor + fPath + "\x1b[0m"
}

func (x *xfg) highlightLine(line string) string {
	h := x.highlighter

	if len(x.extra.searchGrepRe) > 0 {
		for _, re := range x.extra.searchGrepRe {
			line = re.ReplaceAllString(line, h.grepHighlightColor.Sprintf("$1"))
		}
	}

	if x.options.IgnoreCase {
		for _, sgr := range x.extra.searchGrepi {
			line = sgr.ReplaceAllString(line, h.grepHighlightColor.Sprintf("$1"))
		}
	} else {
		for i, sg := range x.options.SearchGrep {
			line = strings.ReplaceAll(line, sg, h.grepHighlighter[i])
		}
	}

	return line
}

func (cli *runner) showResult(x *xfg) error {
	if x.options.Quiet {
		if x.hasMatchedAny() {
			cli.exitCode = exitOK
		} else {
			cli.exitCode = exitErr
		}
		return nil
	}

	// If KeepResultOrder is false, results are already displayed via streaming
	if !x.options.KeepResultOrder {
		cli.exitCode = exitOK
		return nil
	}

	if x.options.NoIndent {
		x.options.Indent = ""
	}

	lf := "\n"
	if x.options.Null {
		lf = "\x00"
	}

	sort.Slice(x.result.paths, func(i, j int) bool { return x.result.paths[i].path < x.result.paths[j].path })

	if cli.isTTY {
		if !x.options.NoColor {
			x.setHighlighter()
		}
		cli.outputForTTY(x, lf)
	} else {
		cli.outputForNonTTY(x, lf)
	}

	cli.exitCode = exitOK

	return nil
}

func (cli *runner) outputForTTY(x *xfg, lf string) error {
	writer := bufio.NewWriter(cli.out)
	for i, p := range x.result.paths {
		if x.options.FilesWithMatches && p.info.IsDir() {
			continue
		}
		out := p.path
		if !x.options.NoColor {
			out = x.highlightPath(out)
		}
		if x.options.ShowMatchCount && !p.info.IsDir() {
			out = out + fmt.Sprintf(":%d", len(p.contents))
		}
		out = out + lf

		if !x.options.ShowMatchCount && !x.options.FilesWithMatches {
			if len(p.contents) > 0 {
				cli.buildContentOutput(x, &out, p.contents, lf)
				if len(x.result.paths)-1 != i {
					out = out + lf
				}
			}
		}

		if x.options.Stats {
			x.cli.stats.AddOutputLC(strings.Count(out, lf))
		}

		if err := xfgutil.Output(writer, out); err != nil {
			return err
		}
	}

	return nil
}

func (cli *runner) buildContentOutput(x *xfg, out *string, contents []line, lf string) error {
	var blc int32 = 0
	for _, line := range contents {
		if !x.options.NoGroupSeparator && x.needToShowGroupSeparator(blc, line.lc) {
			*out = *out + x.options.Indent + x.options.GroupSeparator + lf
		}
		lc := fmt.Sprintf("%d", line.lc)
		if !x.options.NoColor && line.matched {
			lc = x.highlighter.grepHighlightColor.Sprint(lc)
			line.content = x.highlightLine(line.content)
		}
		*out = *out + fmt.Sprintf("%s%s: %s%s", x.options.Indent, lc, line.content, lf)
		blc = line.lc
	}

	return nil
}

func (x *xfg) needToShowGroupSeparator(blc int32, lc int32) bool {
	return (x.options.extra.withAfterContextLines || x.options.extra.withBeforeContextLines) && blc != 0 && lc-blc > 1
}

// streamDisplay displays results as they arrive via channel
func (x *xfg) streamDisplay() {
	if x.options.NoIndent {
		x.options.Indent = ""
	}

	lf := "\n"
	if x.options.Null {
		lf = "\x00"
	}

	if x.cli.isTTY {
		if !x.options.NoColor {
			x.setHighlighter()
		}
		x.cli.streamDisplayTTY(x, lf)
	} else {
		x.cli.streamDisplayNonTTY(x, lf)
	}

	x.streamDone <- true
}

func (cli *runner) streamDisplayTTY(x *xfg, lf string) {
	writer := bufio.NewWriter(cli.out)

	for p := range x.resultChan {
		if x.options.FilesWithMatches && p.info.IsDir() {
			continue
		}

		out := p.path
		if !x.options.NoColor {
			out = x.highlightPath(out)
		}
		if x.options.ShowMatchCount && !p.info.IsDir() {
			out = out + fmt.Sprintf(":%d", len(p.contents))
		}
		out = out + lf

		if !x.options.ShowMatchCount && !x.options.FilesWithMatches {
			if len(p.contents) > 0 {
				var blc int32 = 0
				for _, line := range p.contents {
					if !x.options.NoGroupSeparator && x.needToShowGroupSeparator(blc, line.lc) {
						out = out + x.options.Indent + x.options.GroupSeparator + lf
					}
					lc := fmt.Sprintf("%d", line.lc)
					if !x.options.NoColor && line.matched {
						lc = x.highlighter.grepHighlightColor.Sprint(lc)
						line.content = x.highlightLine(line.content)
					}
					out = out + fmt.Sprintf("%s%s: %s%s", x.options.Indent, lc, line.content, lf)
					blc = line.lc
				}
				out = out + lf
			}
		}

		if x.options.Stats {
			x.cli.stats.AddOutputLC(strings.Count(out, lf))
		}

		if err := xfgutil.Output(writer, out); err != nil {
			break // Error handling: just break on error
		}
	}
}

func (cli *runner) streamDisplayNonTTY(x *xfg, lf string) {
	writer := bufio.NewWriter(cli.out)

	for p := range x.resultChan {
		out := ""
		if len(p.contents) > 0 && !x.options.FilesWithMatches {
			for _, l := range p.contents {
				if l.matched {
					out = out + fmt.Sprintf("%s:%d:%s%s", p.path, l.lc, l.content, lf)
				}
			}
		} else {
			if !x.options.FilesWithMatches || !p.info.IsDir() {
				out = out + fmt.Sprintf("%s%s", p.path, lf)
			}
		}

		if x.options.Stats {
			x.cli.stats.AddOutputLC(strings.Count(out, lf))
		}

		if err := xfgutil.Output(writer, out); err != nil {
			break // Error handling: just break on error
		}
	}
}

func (cli *runner) outputForNonTTY(x *xfg, lf string) error {
	writer := bufio.NewWriter(cli.out)
	for _, p := range x.result.paths {
		out := ""
		if len(p.contents) > 0 && !x.options.FilesWithMatches {
			for _, l := range p.contents {
				if l.matched {
					out = out + fmt.Sprintf("%s:%d:%s%s", p.path, l.lc, l.content, lf)
				}
			}
		} else {
			if !x.options.FilesWithMatches || !p.info.IsDir() {
				out = out + fmt.Sprintf("%s%s", p.path, lf)
			}
		}

		if x.options.Stats {
			x.cli.stats.AddOutputLC(strings.Count(out, lf))
		}

		if err := xfgutil.Output(writer, out); err != nil {
			return err
		}
	}

	return nil
}
