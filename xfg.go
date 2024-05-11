package main

import (
	"io/fs"
	"regexp"
	"sync"

	"github.com/bayashi/xfg/internal/xfgignore"
	"github.com/fatih/color"
)

type line struct {
	lc      int32 // line number
	content string
	matched bool
}

type path struct {
	path     string
	info     fs.DirEntry
	contents []line
}

type result struct {
	mu                  sync.RWMutex
	paths               []path
	outputLC            int // Used on pager. Rough count. Not included group separators.
	alreadyMatchContent bool
}

type highlighter struct {
	pathBaseColor      string
	pathHighlightColor *color.Color
	pathHighlighter    []string
	grepHighlightColor *color.Color
	grepHighlighter    []string
}

type xfgExtra struct {
	searchPathi    []*regexp.Regexp
	searchGrepi    []*regexp.Regexp
	searchPathRe   []*regexp.Regexp
	searchGrepRe   []*regexp.Regexp
	ignoreOptionRe []*regexp.Regexp
	ignoreMatchers xfgignore.Matchers
}

type xfg struct {
	cli         *runner
	options     *options
	highlighter highlighter
	extra       xfgExtra
	result      result
}

func newX(cli *runner, o *options) *xfg {
	o.prepareFromENV()
	o.prepareAliases()
	o.prepareContextLines(cli.isTTY)
	o.prepareRuntimeFlags()

	x := &xfg{
		cli:     cli,
		options: o,
	}

	return x
}
