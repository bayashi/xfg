package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/bayashi/colorpalette"
	"github.com/fatih/color"
	ignore "github.com/sabhiram/go-gitignore"
	"golang.org/x/sync/errgroup"

	"github.com/bayashi/xfg/internal/xfglangxt"
	"github.com/bayashi/xfg/internal/xfgutil"
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
	outputLC            int // Used on pager and stats
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
	o.SearchStart = filepath.Clean(o.SearchStart)

	x := &xfg{
		cli:     cli,
		options: o,
	}

	x.setHighlighter()

	return x
}

func (x *xfg) setHighlighter() {
	o := x.options
	if o.ColorPathBase != "" && colorpalette.Exists(o.ColorPathBase) {
		x.pathBaseColor = fmt.Sprintf("\x1b[%sm", colorpalette.GetCode(o.ColorPathBase))
	} else {
		x.pathBaseColor = fmt.Sprintf("\x1b[%sm", colorpalette.GetCode("yellow"))
	}

	if o.ColorPath != "" && colorpalette.Exists(o.ColorPath) {
		x.pathHighlightColor = colorpalette.Get(o.ColorPath)
	} else {
		x.pathHighlightColor = colorpalette.Get("cyan")
	}
	for _, sp := range o.SearchPath {
		x.pathHighlighter = append(x.pathHighlighter, x.pathHighlightColor.Sprintf(sp))
	}

	if o.ColorContent != "" && colorpalette.Exists(o.ColorContent) {
		x.grepHighlightColor = colorpalette.Get(o.ColorContent)
	} else {
		x.grepHighlightColor = colorpalette.Get("red")
	}
	for _, sg := range o.SearchGrep {
		x.grepHighlighter = append(x.grepHighlighter, x.grepHighlightColor.Sprintf(sg))
	}
}

func (x *xfg) search() error {
	if err := x.preSearch(); err != nil {
		return fmt.Errorf("preSearch() : %w", err)
	}

	if x.options.Stats {
		x.cli.stats.Mark("preSearch")
	}

	eg := new(errgroup.Group)
	x.walkDir(eg, x.options.SearchStart)
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("postMatchPath : %w", err)
	}

	return nil
}

func (x *xfg) walkDir(eg *errgroup.Group, dirPath string) {
	eg.Go(func() error {
		stuff, err := os.ReadDir(dirPath)
		if err != nil {
			if errors.Is(err, fs.ErrPermission) {
				x.cli.putErr(err)
				return nil
			}
			return err
		}

		for _, s := range stuff {
			if !x.options.SearchAll {
				if isDefaultSkipDir(s) || (s.IsDir() && !x.options.Hidden && strings.HasPrefix(s.Name(), ".")) {
					continue // skip all stuff in this dir
				}
			}
			if s.IsDir() {
				x.walkDir(eg, filepath.Join(dirPath, s.Name()))
			}
			if err := x.walkFile(filepath.Join(dirPath, s.Name()), s, eg); err != nil {
				return err
			}
		}

		return nil
	})
}

func (x *xfg) walkFile(fPath string, fInfo fs.DirEntry, eg *errgroup.Group) error {
	if x.options.Stats {
		x.cli.stats.IncrWalkedPaths()
	}

	if x.options.Quiet && x.hasMatchedAny() {
		return nil // already match. skip after all
	}

	if !x.options.SearchAll && len(x.options.Ext) > 0 && !x.isMatchExt(fInfo, x.options.Ext) {
		return nil
	}

	if !x.options.SearchAll && len(x.options.Lang) > 0 && !x.isLangFile(fInfo) {
		return nil
	}

	if x.isSkippable(fPath, fInfo) {
		return nil
	}

	if x.options.Stats {
		x.cli.stats.IncrWalkedContents()
	}

	eg.Go(func() error {
		return x.postMatchPath(fPath, fInfo)
	})

	return nil
}

func (x *xfg) preSearch() error {
	if !x.options.SkipGitIgnore {
		x.gitignore = prepareGitIgnore(x.cli.homeDir, x.options.SearchStart)
	}

	if !x.options.SkipXfgIgnore {
		x.xfgignore = prepareXfgIgnore(x.cli.homeDir, x.options.XfgIgnoreFile)
	}

	if x.options.IgnoreCase {
		if err := x.prepareRe(); err != nil {
			return err
		}
	}

	if len(x.options.SearchPathRe) > 0 {
		if searchPathRe, err := xfgutil.CompileRegexps(x.options.SearchPathRe, !x.options.NotWordBoundary); err != nil {
			return err
		} else {
			x.searchPathRe = searchPathRe
		}
	}

	if len(x.options.SearchGrepRe) > 0 {
		if searchGrepRe, err := xfgutil.CompileRegexps(x.options.SearchGrepRe, !x.options.NotWordBoundary); err != nil {
			return err
		} else {
			x.searchGrepRe = searchGrepRe
		}
	}

	return nil
}

func (x *xfg) prepareRe() error {
	if searchPathi, err := xfgutil.CompileRegexpsIgnoreCase(x.options.SearchPath); err != nil {
		return err
	} else {
		x.searchPathi = searchPathi
	}

	if len(x.options.SearchGrep) > 0 {
		if searchGrepi, err := xfgutil.CompileRegexpsIgnoreCase(x.options.SearchGrep); err != nil {
			return err
		} else {
			x.searchGrepi = searchGrepi
		}
	}

	if len(x.options.Ignore) > 0 {
		if ignoreRe, err := xfgutil.CompileRegexpsIgnoreCase(x.options.Ignore); err != nil {
			return err
		} else {
			x.ignoreRe = ignoreRe
		}
	}

	return nil
}

func (x *xfg) isMatchExt(fInfo fs.DirEntry, extensions []string) bool {
	for _, ext := range extensions {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		if strings.HasSuffix(fInfo.Name(), ext) {
			return true
		}
	}

	return false
}

func (x *xfg) isLangFile(fInfo fs.DirEntry) bool {
	for _, l := range x.options.Lang {
		if xfglangxt.IsLangFile(l, fInfo.Name()) {
			return true
		}
	}

	return false
}

func (x *xfg) isSkippable(fPath string, fInfo fs.DirEntry) bool {
	if fInfo.IsDir() && x.options.onlyMatchContent {
		return true // Just not pick up only this dir path. It will be searched files and directories in this dir.
	}

	if x.isIgnorePath(fPath) {
		return true
	}

	if !x.options.SearchAll {
		if isDefaultSkipFile(fInfo) ||
			(!x.options.Hidden && strings.HasPrefix(fInfo.Name(), ".")) ||
			x.isSkippableByIgnoreFile(fPath) {
			return true
		}
	}

	return x.canSkipPath(fPath, fInfo)
}

func (x *xfg) isSkippableByIgnoreFile(fPath string) bool {
	if x.gitignore != nil && x.gitignore.MatchesPath(fPath) {
		return true // skip a file by .gitignore
	}
	if x.xfgignore != nil && x.xfgignore.MatchesPath(fPath) {
		return true // skip a file by .xfgignore
	}

	return false
}

func (x *xfg) isIgnorePath(fPath string) bool {
	if x.options.IgnoreCase {
		for _, re := range x.ignoreRe {
			if isMatchRegexp(fPath, re) {
				return true // ignore
			}
		}
	} else {
		for _, i := range x.options.Ignore {
			if isMatch(fPath, i) {
				return true // ignore
			}
		}
	}

	return false
}

func (x *xfg) canSkipPath(fPath string, fInfo fs.DirEntry) bool {
	if x.options.SearchOnlyName {
		return x._canSkipPath(fInfo.Name())
	}

	return x._canSkipPath(fPath)
}

func (x *xfg) _canSkipPath(fPath string) bool {
	if x.options.IgnoreCase && len(x.searchPathi) > 0 {
		for _, spr := range x.searchPathi {
			if !isMatchRegexp(fPath, spr) {
				return true // OK, skip
			}
		}
	} else if len(x.options.SearchPath) > 0 {
		for _, sp := range x.options.SearchPath {
			if !isMatch(fPath, sp) {
				return true // OK, skip
			}
		}
	}

	if len(x.searchPathRe) > 0 {
		for _, re := range x.searchPathRe {
			if !isMatchRegexp(fPath, re) {
				return true // OK, skip
			}
		}
	}

	return false // match all, cannot skip
}

func (x *xfg) postMatchPath(fPath string, fInfo fs.DirEntry) (err error) {
	matchedPath := path{
		info: fInfo,
	}

	if (len(x.options.SearchGrep) > 0 || len(x.searchGrepRe) > 0) && isRegularFile(fInfo) {
		matchedPath.contents, err = x.scanFile(fPath)
		if err != nil {
			return fmt.Errorf("scanFile() : %w", err)
		}
	}

	if x.options.onlyMatchContent && len(matchedPath.contents) == 0 {
		return nil // not pick up
	}

	if x.options.Abs {
		absPath, err := filepath.Abs(fPath)
		if err != nil {
			return fmt.Errorf("failed to get abs path of `%s` : %w", fPath, err)
		}
		fPath = absPath
	}

	if fInfo.IsDir() {
		fPath = fPath + string(filepath.Separator)
	}

	if !x.options.NoColor {
		fPath = x.highlightPath(fPath)
	}

	matchedPath.path = fPath

	x.result.mu.Lock()
	x.result.paths = append(x.result.paths, matchedPath)
	x.result.outputLC = x.result.outputLC + len(matchedPath.contents) + 1
	x.result.mu.Unlock()

	return nil
}

func (x *xfg) scanFile(fPath string) ([]line, error) {
	if x.options.Stats {
		x.cli.stats.Lock()
		x.cli.stats.IncrScannedFile()
		x.cli.stats.Unlock()
	}

	fh, err := os.Open(fPath)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			x.cli.putErr(err)
			return nil, nil
		}
		return nil, fmt.Errorf("path `%s` : %w", fPath, err)
	}
	defer fh.Close()

	isBinary, err := isBinaryFile(fh)
	if err != nil {
		return nil, fmt.Errorf("path `%s` : %w", fPath, err)
	}
	if isBinary {
		return nil, nil
	}

	if _, err := fh.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("could not seek `%s` : %w", fPath, err)
	}

	matchedContents, err := x.scanContent(bufio.NewScanner(fh), fPath)
	if err != nil {
		return nil, fmt.Errorf("scanContent() `%s` : %w", fPath, err)
	}

	return matchedContents, nil
}

func (x *xfg) highlightPath(fPath string) string {
	if len(x.searchPathRe) > 0 {
		for _, re := range x.searchPathRe {
			fPath = re.ReplaceAllString(fPath, x.pathHighlightColor.Sprintf("$1")+x.pathBaseColor)
		}
	}

	if x.options.IgnoreCase {
		for _, spr := range x.searchPathi {
			fPath = spr.ReplaceAllString(fPath, x.pathHighlightColor.Sprintf("$1")+x.pathBaseColor)
		}
	} else {
		for i, sp := range x.options.SearchPath {
			fPath = strings.ReplaceAll(fPath, sp, x.pathHighlighter[i]+x.pathBaseColor)
		}
	}

	return x.pathBaseColor + fPath + "\x1b[0m"
}

type scanFile struct {
	lc     int32  // line count
	l      string // line text
	blines []line // slice for before lines
	aline  uint32 // the count for after lines

	matchedContents []line // result
}

func (x *xfg) scanContent(scanner *bufio.Scanner, fPath string) ([]line, error) {
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

		if x.options.MaxMatchCount != 0 && int(x.options.MaxMatchCount) <= len(gf.matchedContents) {
			break
		}
	}

	if x.options.Stats {
		x.cli.stats.Lock()
		x.cli.stats.IncrScannedLC(int(gf.lc))
		x.cli.stats.Unlock()
	}

	if x.options.Quiet && !x.result.alreadyMatchContent && len(gf.matchedContents) > 0 {
		x.result.alreadyMatchContent = true
	}

	return gf.matchedContents, nil
}

func (x *xfg) isMatchLine(line string) bool {
	if x.options.IgnoreCase && len(x.searchGrepi) > 0 {
		for _, sgr := range x.searchGrepi {
			if !isMatchRegexp(line, sgr) {
				return false
			}
		}
	} else {
		if len(x.options.SearchGrep) > 0 {
			for _, sg := range x.options.SearchGrep {
				if !isMatch(line, sg) {
					return false
				}
			}
		}
	}

	if len(x.searchGrepRe) > 0 {
		for _, re := range x.searchGrepRe {
			if !isMatchRegexp(line, re) {
				return false
			}
		}
	}

	return true // OK, match all
}

func (x *xfg) highlightLine(gf *scanFile) {
	if len(x.searchGrepRe) > 0 {
		for _, re := range x.searchGrepRe {
			gf.l = re.ReplaceAllString(gf.l, x.grepHighlightColor.Sprintf("$1"))
		}
	}

	if x.options.IgnoreCase {
		for _, sgr := range x.searchGrepi {
			gf.l = sgr.ReplaceAllString(gf.l, x.grepHighlightColor.Sprintf("$1"))
		}
	} else {
		for i, sg := range x.options.SearchGrep {
			gf.l = strings.ReplaceAll(gf.l, sg, x.grepHighlighter[i])
		}
	}
}

func (x *xfg) processContentLine(gf *scanFile) {
	if x.isMatchLine(gf.l) {
		if !x.options.ShowMatchCount && x.options.withBeforeContextLines {
			for _, bl := range gf.blines {
				if bl.lc == 0 {
					continue // skip
				}
				gf.matchedContents = append(gf.matchedContents, bl)
			}
			gf.blines = make([]line, x.options.actualBeforeContextLines)
		}

		if x.options.ShowMatchCount {
			gf.l = ""
		} else if !x.options.NoColor {
			x.highlightLine(gf)
		}

		x.optimizeLine(gf)
		gf.matchedContents = append(gf.matchedContents, line{lc: gf.lc, content: gf.l, matched: true})

		if !x.options.ShowMatchCount && x.options.withAfterContextLines {
			gf.aline = x.options.actualAfterContextLines // start countdown for `aline`
		}
	} else {
		if !x.options.ShowMatchCount {
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
	x.result.mu.RLock()
	defer x.result.mu.RUnlock()
	if (len(x.options.SearchGrep) == 0 && len(x.result.paths) > 0) ||
		(len(x.options.SearchGrep) > 0 && len(x.result.paths) > 0 && x.result.alreadyMatchContent) {
		return true // already match
	}

	return false
}

func (x *xfg) optimizeLine(gf *scanFile) {
	if x.options.MaxColumns > 0 && len(gf.l) > int(x.options.MaxColumns) {
		gf.l = gf.l[:x.options.MaxColumns]
	}
}
