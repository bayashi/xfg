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

	if cli.isTTY {
		cli.outputForTTY(x)
	} else {
		cli.outputForNonTTY(x)
	}

	cli.exitCode = exitOK

	return nil
}

func (cli *runner) outputForTTY(x *xfg) error {
	writer := bufio.NewWriter(cli.out)
	for _, p := range x.result {
		out := p.path
		if x.options.ShowMatchCount && !p.info.IsDir() {
			out = out + fmt.Sprintf(":%d", len(p.contents))
		}
		out = out + "\n"

		if !x.options.ShowMatchCount {
			if len(p.contents) > 0 {
				cli.buildContentOutput(x, &out, p.contents)
			}
			if x.options.Relax && len(p.contents) > 0 {
				out = out + "\n"
			}
		}
		if err := output(writer, out); err != nil {
			return err
		}
	}

	return nil
}

func (cli *runner) buildContentOutput(x *xfg, out *string, contents []line) error {
	var blc int32 = 0
	for _, line := range contents {
		if !x.options.NoGroupSeparator && x.needToShowGroupSeparator(blc, line.lc) {
			*out = *out + x.options.Indent + x.options.GroupSeparator + "\n"
		}
		lc := fmt.Sprintf("%d", line.lc)
		if !x.options.NoColor && line.matched {
			lc = x.grepHighlightColor.Sprint(lc)
		}
		*out = *out + fmt.Sprintf("%s%s: %s\n", x.options.Indent, lc, line.content)
		blc = line.lc
	}

	return nil
}

func (x *xfg) needToShowGroupSeparator(blc int32, lc int32) bool {
	return (x.options.withAfterContextLines || x.options.withBeforeContextLines) && blc != 0 && lc-blc > 1
}

func (cli *runner) outputForNonTTY(x *xfg) error {
	writer := bufio.NewWriter(cli.out)
	for _, p := range x.result {
		out := ""
		if p.info.IsDir() {
			out = fmt.Sprintf("%s\n", p.path)
		} else {
			for _, l := range p.contents {
				if l.matched {
					out = out + fmt.Sprintf("%s:%d:%s\n", p.path, l.lc, l.content)
				}
			}
		}

		if err := output(writer, out); err != nil {
			return err
		}
	}

	return nil
}
