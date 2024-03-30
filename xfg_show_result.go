package main

import (
	"bufio"
	"fmt"
	"io"
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

func (x *xfg) showResult(w io.Writer) error {
	if x.options.noIndent {
		x.options.indent = ""
	}

	writer := bufio.NewWriter(w)
	for _, p := range x.result {
		out := p.path
		if x.options.showMatchCount && !p.info.IsDir() {
			out = out + fmt.Sprintf(":%d", len(p.contents))
		}
		out = out + "\n"

		if !x.options.showMatchCount {
			if len(p.contents) > 0 {
				x.buildContentOutput(&out, p.contents)
			}
			if x.options.relax && len(p.contents) > 0 {
				out = out + "\n"
			}
		}
		if err := output(writer, out); err != nil {
			return err
		}
	}

	return nil
}

func (x *xfg) buildContentOutput(out *string, contents []line) error {
	var blc int32 = 0
	for _, line := range contents {
		if x.needToShowGroupSeparator(blc, line.lc) {
			*out = *out + x.options.indent + x.options.groupSeparator + "\n"
		}
		lc := fmt.Sprintf("%d", line.lc)
		if !x.options.noColor && line.matched {
			lc = x.grepHighlightColor.Sprint(lc)
		}
		*out = *out + fmt.Sprintf("%s%s: %s\n", x.options.indent, lc, line.content)
		blc = line.lc
	}

	return nil
}

func (x *xfg) needToShowGroupSeparator(blc int32, lc int32) bool {
	return (x.options.withAfterContextLines || x.options.withBeforeContextLines) && blc != 0 && lc-blc > 1
}
