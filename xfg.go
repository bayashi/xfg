package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bayashi/colorpalette"
	"github.com/fatih/color"
	ignore "github.com/sabhiram/go-gitignore"
)

type line struct {
	lc      int32
	content string
	matched bool
}

type path struct {
	path     string
	info     os.FileInfo
	contents []line
}

type xfg struct {
	options *options

	pathHighlightColor *color.Color
	pathHighlighter    []string
	grepHighlightColor *color.Color
	grepHighlighter    []string

	searchPathRe []*regexp.Regexp
	searchGrepRe []*regexp.Regexp
	ignoreRe     []*regexp.Regexp
	gitignore    *ignore.GitIgnore
	xfgignore    *ignore.GitIgnore

	result             []path
	resultLines        int
	resultMatchContent bool
}

func (x *xfg) setHighlighter() {
	o := x.options
	if o.colorPath != "" && colorpalette.Exists(o.colorPath) {
		x.pathHighlightColor = colorpalette.Get(o.colorPath)
	} else {
		x.pathHighlightColor = colorpalette.Get("cyan")
	}
	for _, sp := range o.searchPath {
		x.pathHighlighter = append(x.pathHighlighter, x.pathHighlightColor.Sprintf(sp))
	}

	if o.colorContent != "" && colorpalette.Exists(o.colorContent) {
		x.grepHighlightColor = colorpalette.Get(o.colorContent)
	} else {
		x.grepHighlightColor = colorpalette.Get("red")
	}
	for _, sg := range o.searchGrep {
		x.grepHighlighter = append(x.grepHighlighter, x.grepHighlightColor.Sprintf(sg))
	}
}

func newX(o *options) *xfg {
	x := &xfg{
		options: o,
	}

	x.setHighlighter()

	return x
}

func (x *xfg) preSearch() error {
	if err := validateStartPath(x.options.searchStart); err != nil {
		return err
	}

	if !x.options.skipGitIgnore {
		x.gitignore = compileGitIgnore(x.options.searchStart)
	}

	if !x.options.skipXfgIgnore {
		x.xfgignore = compileXfgIgnore(x.options.xfgIgnoreFile)
	}

	if x.options.ignoreCase {
		for _, sp := range x.options.searchPath {
			searchPathRe, err := regexp.Compile("(?i)(" + regexp.QuoteMeta(sp) + ")")
			if err != nil {
				return err
			}
			x.searchPathRe = append(x.searchPathRe, searchPathRe)
		}

		if len(x.options.searchGrep) > 0 {
			for _, sg := range x.options.searchGrep {
				searchGrepRe, err := regexp.Compile("(?i)(" + regexp.QuoteMeta(sg) + ")")
				if err != nil {
					return err
				}
				x.searchGrepRe = append(x.searchGrepRe, searchGrepRe)
			}
		}

		if len(x.options.ignore) > 0 {
			for _, i := range x.options.ignore {
				ignoreRe, err := regexp.Compile(`(?i)` + regexp.QuoteMeta(i))
				if err != nil {
					return err
				}
				x.ignoreRe = append(x.ignoreRe, ignoreRe)
			}
		}
	}

	return nil
}

func (x *xfg) search() error {
	if err := x.preSearch(); err != nil {
		return fmt.Errorf("error in preSearch: %w", err)
	}

	walkErr := filepath.Walk(x.options.searchStart, func(fPath string, fInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("something went wrong within path `%s` at `%s`: %w", x.options.searchStart, fPath, err)
		}

		if x.options.quiet {
			if x.hasMatchedAny() {
				return nil // already match. skip after all
			}
		}

		return x.walker(fPath, fInfo)
	})
	if walkErr != nil {
		return fmt.Errorf("failed to walk: %w", walkErr)
	}

	return nil
}

func (x *xfg) walker(fPath string, fInfo os.FileInfo) error {
	if x.isIgnorePath(fPath) {
		return nil // skip by --ignore option
	}

	if !x.options.searchAll {
		if fInfo.IsDir() && fInfo.Name() == ".git" {
			return filepath.SkipDir // not search for .git directory
		}
	}

	if x.canSkip(fPath, fInfo) {
		return nil // skip
	}

	x.onMatchPath(fPath, fInfo)

	return nil
}

func (x *xfg) isIgnorePath(fPath string) bool {
	if x.options.ignoreCase {
		for _, re := range x.ignoreRe {
			if isMatchRegexp(fPath, re) {
				return true // ignore
			}
		}
	} else {
		for _, i := range x.options.ignore {
			if isMatch(fPath, i) {
				return true // ignore
			}
		}
	}

	return false
}

func (x *xfg) canSkip(fPath string, fInfo fs.FileInfo) bool {
	if !x.options.searchAll {
		if canSkipStuff(fInfo) {
			return true // not pick .gitkeep file
		} else if !x.options.hidden && strings.HasPrefix(fInfo.Name(), ".") {
			return true // skip dot-file/dir
		}
	}

	if !x.options.searchAll {
		if x.gitignore != nil && x.gitignore.MatchesPath(fPath) {
			return true // skip a file by .gitignore
		}
		if x.xfgignore != nil && x.xfgignore.MatchesPath(fPath) {
			return true // skip a file by .xfgignore
		}
	}

	if fInfo.IsDir() && x.options.onlyMatchContent {
		return true // not pick up
	}

	return x.canSkipPath(fPath)
}

func (x *xfg) canSkipPath(fPath string) bool {
	if x.options.ignoreCase {
		for _, spr := range x.searchPathRe {
			if !isMatchRegexp(fPath, spr) {
				return true // OK, skip
			}
		}
	} else {
		for _, sp := range x.options.searchPath {
			if !isMatch(fPath, sp) {
				return true // OK, skip
			}
		}
	}

	return false // match all, cannot skip
}

func (x *xfg) onMatchPath(fPath string, fInfo fs.FileInfo) (err error) {
	matchedPath := path{
		info: fInfo,
	}

	if len(x.options.searchGrep) > 0 && isRegularFile(fInfo) {
		matchedPath.contents, err = x.checkFile(fPath)
		if err != nil {
			return fmt.Errorf("error during grep: %w", err)
		}
		if x.options.onlyMatchContent && len(matchedPath.contents) == 0 {
			return nil // not pick up
		}
	}

	if x.options.abs {
		absPath, err := filepath.Abs(fPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path from `%s`: %w", fPath, err)
		}
		fPath = absPath
	}

	matchedPath.path = fPath
	if !x.options.noColor {
		if x.options.ignoreCase {
			for _, spr := range x.searchPathRe {
				matchedPath.path = spr.ReplaceAllString(matchedPath.path, x.pathHighlightColor.Sprintf("$1"))
			}
		} else {
			for i, sp := range x.options.searchPath {
				matchedPath.path = strings.ReplaceAll(matchedPath.path, sp, x.pathHighlighter[i])
			}
		}
	}

	if fInfo.IsDir() {
		matchedPath.path = matchedPath.path + string(filepath.Separator)
	}

	x.result = append(x.result, matchedPath)
	x.resultLines = x.resultLines + len(matchedPath.contents) + 1

	return nil
}

func (x *xfg) checkFile(fPath string) ([]line, error) {
	fh, err := os.Open(fPath)
	if err != nil {
		return nil, fmt.Errorf("could not open file `%s`: %w", fPath, err)
	}
	defer fh.Close()

	isBinary, err := isBinaryFile(fh)
	if err != nil {
		return nil, fmt.Errorf("error during isBinary file `%s`: %w", fPath, err)
	}
	if isBinary {
		return nil, nil
	}

	if _, err := fh.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("could not seek `%s`: %w", fPath, err)
	}

	matchedContents, err := x.scanFile(bufio.NewScanner(fh), fPath)
	if err != nil {
		return nil, fmt.Errorf("could not grepFile `%s`: %w", fPath, err)
	}

	return matchedContents, nil
}

type scanFile struct {
	lc     int32  // line count
	l      string // line text
	blines []line // slice for before lines
	aline  uint32 // the count for after lines

	matchedContents []line // result
}

func (x *xfg) scanFile(scanner *bufio.Scanner, fPath string) ([]line, error) {
	gf := &scanFile{
		lc:     0,
		blines: make([]line, x.options.actualBeforeContextLines),
	}

	for scanner.Scan() {
		gf.lc++
		gf.l = scanner.Text()
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("could not scan file `%s` line %d: %w", fPath, gf.lc, err)
		}

		x.processContentLine(gf)

		if x.options.maxMatchCount != 0 && int(x.options.maxMatchCount) <= len(gf.matchedContents) {
			break
		}
	}

	if x.options.quiet && len(gf.matchedContents) > 0 {
		x.resultMatchContent = true
	}

	return gf.matchedContents, nil
}

func (x *xfg) isMatchLine(line string) bool {
	if x.options.ignoreCase {
		for _, sgr := range x.searchGrepRe {
			if !isMatchRegexp(line, sgr) {
				return false
			}
		}
		return true // OK, match all
	} else {
		for _, sg := range x.options.searchGrep {
			if !isMatch(line, sg) {
				return false
			}
		}
		return true // OK, match all
	}
}

func (x *xfg) processContentLine(gf *scanFile) {
	if x.isMatchLine(gf.l) {
		if !x.options.showMatchCount && x.options.withBeforeContextLines {
			for _, bl := range gf.blines {
				if bl.lc == 0 {
					continue // skip
				}
				gf.matchedContents = append(gf.matchedContents, bl)
			}
			gf.blines = make([]line, x.options.actualBeforeContextLines)
		}

		if x.options.showMatchCount {
			gf.l = ""
		} else if !x.options.noColor {
			if x.options.ignoreCase {
				for _, sgr := range x.searchGrepRe {
					gf.l = sgr.ReplaceAllString(gf.l, x.grepHighlightColor.Sprintf("$1"))
				}
			} else {
				for i, sg := range x.options.searchGrep {
					gf.l = strings.ReplaceAll(gf.l, sg, x.grepHighlighter[i])
				}
			}
		}

		x.optimizeLine(gf)
		gf.matchedContents = append(gf.matchedContents, line{lc: gf.lc, content: gf.l, matched: true})

		if !x.options.showMatchCount && x.options.withAfterContextLines {
			gf.aline = x.options.actualAfterContextLines // start countdown for `aline`
		}
	} else {
		if !x.options.showMatchCount {
			if x.options.withAfterContextLines && gf.aline > 0 {
				gf.aline--
				x.optimizeLine(gf)
				gf.matchedContents = append(gf.matchedContents, line{lc: gf.lc, content: gf.l})
			} else if x.options.withBeforeContextLines {
				// rotate blines
				// join "2nd to last elements of `blines`" and "current `line`"
				x.optimizeLine(gf)
				gf.blines = append(gf.blines[1:], line{lc: gf.lc, content: gf.l})
			}
		}
	}
}

func (x *xfg) hasMatchedAny() bool {
	if (len(x.options.searchGrep) == 0 && len(x.result) > 0) ||
		(len(x.options.searchGrep) > 0 && len(x.result) > 0 && x.resultMatchContent) {
		return true // already match
	}

	return false
}

func (x *xfg) optimizeLine(gf *scanFile) {
	if x.options.maxColumns > 0 && len(gf.l) > int(x.options.maxColumns) {
		gf.l = gf.l[:x.options.maxColumns]
	}
}
