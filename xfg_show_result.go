package main

import (
	"bufio"
	"fmt"
)

func output(writer *bufio.Writer, out string) error {
	if _, err := fmt.Fprint(writer, out); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
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

	if x.options.NoIndent {
		x.options.Indent = ""
	}

	lf := "\n"
	if x.options.Null {
		lf = "\x00"
	}

	if cli.isTTY {
		cli.outputForTTY(x, lf)
	} else {
		cli.outputForNonTTY(x, lf)
	}

	cli.exitCode = exitOK

	return nil
}

func (cli *runner) outputForTTY(x *xfg, lf string) error {
	writer := bufio.NewWriter(cli.out)
	for _, p := range x.result.paths {
		if x.options.FilesWithMatches && p.info.IsDir() {
			continue
		}
		out := p.path
		if x.options.ShowMatchCount && !p.info.IsDir() {
			out = out + fmt.Sprintf(":%d", len(p.contents))
		}
		out = out + "\n"

		if !x.options.ShowMatchCount && !x.options.FilesWithMatches {
			if len(p.contents) > 0 {
				cli.buildContentOutput(x, &out, p.contents, lf)
			}
			if x.options.Relax && len(p.contents) > 0 {
				out = out + lf
			}
		}
		if err := output(writer, out); err != nil {
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
			lc = x.grepHighlightColor.Sprint(lc)
		}
		*out = *out + fmt.Sprintf("%s%s: %s%s", x.options.Indent, lc, line.content, lf)
		blc = line.lc
	}

	return nil
}

func (x *xfg) needToShowGroupSeparator(blc int32, lc int32) bool {
	return (x.options.withAfterContextLines || x.options.withBeforeContextLines) && blc != 0 && lc-blc > 1
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

		if err := output(writer, out); err != nil {
			return err
		}
	}

	return nil
}
