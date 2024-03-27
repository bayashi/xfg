package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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

	pathHighlitColor *color.Color
	pathHighlighter  string
	grepHighlitColor *color.Color
	grepHighlighter  string

	searchPathRe *regexp.Regexp
	searchGrepRe *regexp.Regexp
	ignoreRe     []*regexp.Regexp

	result []path
}

func newX(o *options, pathHighlightColor *color.Color, grepHighlightColor *color.Color) *xfg {
	x := &xfg{
		options: o,
	}
	if pathHighlightColor != nil {
		x.pathHighlitColor = pathHighlightColor
		x.pathHighlighter = pathHighlightColor.Sprintf(x.options.searchPath)
	}
	if grepHighlightColor != nil {
		x.grepHighlitColor = grepHighlightColor
		x.grepHighlighter = grepHighlightColor.Sprintf(x.options.searchGrep)
	}

	return x
}

type walkerArg struct {
	path      string
	info      fs.FileInfo
	gitignore *ignore.GitIgnore
}

func (x *xfg) search() error {
	if err := validateStartPath(x.options.searchStart); err != nil {
		return err
	}

	var gitignore *ignore.GitIgnore
	if !x.options.skipGitIgnore {
		gitignore = compileGitIgnore(x.options.searchStart)
	}

	if x.options.ignoreCase {
		searchPathRe, err := regexp.Compile("(?i)(" + regexp.QuoteMeta(x.options.searchPath) + ")")
		if err != nil {
			return err
		}
		x.searchPathRe = searchPathRe
		if x.options.searchGrep != "" {
			searchGrepRe, err := regexp.Compile("(?i)(" + regexp.QuoteMeta(x.options.searchGrep) + ")")
			if err != nil {
				return err
			}
			x.searchGrepRe = searchGrepRe
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

	walkErr := filepath.Walk(x.options.searchStart, func(fPath string, fInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("something went wrong within path `%s` at `%s`: %w", x.options.searchStart, fPath, err)
		}

		return x.walker(&walkerArg{
			path:      fPath,
			info:      fInfo,
			gitignore: gitignore,
		})
	})
	if walkErr != nil {
		return fmt.Errorf("failed to walk: %w", walkErr)
	}

	return nil
}

func (x *xfg) walker(wa *walkerArg) error {
	fPath, fInfo := wa.path, wa.info

	if x.isIgnorePath(fPath) {
		return nil // skip by --ignore option
	}

	if !x.options.searchAll {
		if fInfo.IsDir() && fInfo.Name() == ".git" {
			return filepath.SkipDir // not search for .git directory
		}
	}

	if x.canSkip(fPath, fInfo, wa.gitignore) {
		return nil // skip
	}

	x.onMatchPath(fPath, fInfo)

	return nil
}

func (x *xfg) isIgnorePath(fPath string) bool {
	if x.options.ignoreCase {
		for _, re := range x.ignoreRe {
			if isMatchRegexp(fPath, re) {
				return true // skip
			}
		}
	} else {
		for _, i := range x.options.ignore {
			if isMatch(fPath, i) {
				return true // skip
			}
		}
	}

	return false
}

func (x *xfg) canSkip(fPath string, fInfo fs.FileInfo, gitignore *ignore.GitIgnore) bool {
	if !x.options.searchAll {
		if !fInfo.IsDir() && (fInfo.Name() == ".gitkeep" || strings.HasSuffix(fInfo.Name(), ".min.js")) {
			return true // not pick .gitkeep file
		} else if !x.options.hidden && strings.HasPrefix(fInfo.Name(), ".") {
			return true // skip dot-file
		}
	}

	if !x.options.searchAll && gitignore != nil && gitignore.MatchesPath(fPath) {
		return true // skip a file by .gitignore
	}

	if fInfo.IsDir() {
		if x.options.onlyMatch {
			return true // not pick up
		}
	}

	if x.options.ignoreCase {
		return !isMatchRegexp(fPath, x.searchPathRe)
	} else {
		return !isMatch(fPath, x.options.searchPath)
	}
}

func (x *xfg) onMatchPath(fPath string, fInfo fs.FileInfo) (err error) {
	matchedPath := path{
		info: fInfo,
	}

	if x.options.searchGrep != "" && isRegularFile(fInfo) {
		matchedPath.contents, err = x.checkFile(fPath)
		if err != nil {
			return fmt.Errorf("error during grep: %w", err)
		}
		if x.options.onlyMatch && len(matchedPath.contents) == 0 {
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

	if x.options.noColor {
		matchedPath.path = fPath
	} else {
		if x.options.ignoreCase {
			matchedPath.path = x.searchPathRe.ReplaceAllString(fPath, x.pathHighlitColor.Sprintf("$1"))
		} else {
			matchedPath.path = strings.ReplaceAll(fPath, x.options.searchPath, x.pathHighlighter)
		}
	}

	if fInfo.IsDir() {
		matchedPath.path = matchedPath.path + string(filepath.Separator)
	}

	x.result = append(x.result, matchedPath)

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
	lc               int32  // line count
	l                string // line text
	blines           []line // slice for before lines
	aline            uint32 // the count for after lines
	withContextLines bool

	matchedContents []line // result
}

func (x *xfg) scanFile(scanner *bufio.Scanner, fPath string) ([]line, error) {
	gf := &scanFile{
		lc:               0,
		blines:           make([]line, x.options.contextLines),
		withContextLines: x.options.contextLines > 0,
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

	return gf.matchedContents, nil
}

func (x *xfg) isMatchLine(line string) bool {
	if x.options.ignoreCase {
		return isMatchRegexp(line, x.searchGrepRe)
	} else {
		return isMatch(line, x.options.searchGrep)
	}
}

func (x *xfg) processContentLine(gf *scanFile) {
	if x.isMatchLine(gf.l) {
		if !x.options.showMatchCount && gf.withContextLines {
			for _, bl := range gf.blines {
				if bl.lc == 0 {
					continue // skip
				}
				gf.matchedContents = append(gf.matchedContents, bl)
			}
			gf.blines = make([]line, x.options.contextLines)
		}

		if x.options.showMatchCount {
			gf.l = ""
		} else if !x.options.noColor {
			if x.options.ignoreCase {
				gf.l = x.searchGrepRe.ReplaceAllString(gf.l, x.grepHighlitColor.Sprintf("$1"))
			} else {
				gf.l = strings.ReplaceAll(gf.l, x.options.searchGrep, x.grepHighlighter)
			}
		}

		gf.matchedContents = append(gf.matchedContents, line{lc: gf.lc, content: gf.l, matched: true})

		if !x.options.showMatchCount && gf.withContextLines {
			gf.aline = x.options.contextLines // start countdown for `aline`
		}
	} else {
		if !x.options.showMatchCount && gf.withContextLines {
			if gf.aline > 0 {
				gf.aline--
				gf.matchedContents = append(gf.matchedContents, line{lc: gf.lc, content: gf.l})
			} else {
				// rotate blines
				// join "2nd to last elements of `blines`" and "current `line`"
				gf.blines = append(gf.blines[1:], line{lc: gf.lc, content: gf.l})
			}
		}
	}
}
