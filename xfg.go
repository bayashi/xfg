package main

import (
	"io/fs"
	"regexp"
	"sync"

	"github.com/fatih/color"
	ignore "github.com/sabhiram/go-gitignore"
)

const (
	XFG_NO_COLOR_ENV_KEY = "NO_COLOR"
	XFG_RC_ENV_KEY       = "XFG_RC_FILE_PATH"
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

type xfg struct {
	cli     *runner
	options *options

	pathBaseColor      string
	pathHighlightColor *color.Color
	pathHighlighter    []string
	grepHighlightColor *color.Color
	grepHighlighter    []string

	searchPathi  []*regexp.Regexp
	searchGrepi  []*regexp.Regexp
	searchPathRe []*regexp.Regexp
	searchGrepRe []*regexp.Regexp
	ignoreRe     []*regexp.Regexp
	gitignore    *ignore.GitIgnore
	xfgignore    *ignore.GitIgnore

	result result
}

func newX(cli *runner, o *options) *xfg {
	o.prepareFromENV()
	o.prepareAliases()
	o.prepareContextLines(cli.isTTY)

	x := &xfg{
		cli:     cli,
		options: o,
	}

	return x
}
