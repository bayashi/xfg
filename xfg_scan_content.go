package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type scanFile struct {
	lc     int32  // line count
	l      string // line text
	blines []line // slice for before lines
	aline  uint32 // the count for after lines

	matchedContents []line // result
}

func (x *xfg) postMatchPath(fPath string, fInfo fs.DirEntry) (err error) {
	matchedPath := path{
		info: fInfo,
	}

	if (len(x.options.SearchGrep) > 0 || len(x.extra.searchGrepRe) > 0) && isRegularFile(fInfo) {
		matchedPath.contents, err = x.scanFile(fPath)
		if err != nil {
			return fmt.Errorf("scanFile() : %w", err)
		}
	}

	if x.options.extra.onlyMatchContent && len(matchedPath.contents) == 0 {
		return nil // not pick up
	}

	return x.postScanFile(fPath, fInfo, matchedPath)
}

func (x *xfg) scanFile(fPath string) ([]line, error) {
	if x.options.Stats {
		x.cli.stats.IncrScannedFile()
	}

	fh, err := os.Open(fPath)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			if !x.options.IgnorePermissionError {
				x.cli.putErr(err)
			}
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

func (x *xfg) postScanFile(fPath string, fInfo fs.DirEntry, matchedPath path) error {
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

	matchedPath.path = fPath

	x.result.mu.Lock()
	x.result.paths = append(x.result.paths, matchedPath)
	x.result.outputLC = x.result.outputLC + len(matchedPath.contents) + 1
	x.result.mu.Unlock()

	if x.options.Stats {
		x.cli.stats.AddPickedLC(len(matchedPath.contents))
	}

	return nil
}

func (x *xfg) scanContent(scanner *bufio.Scanner, fPath string) ([]line, error) {
	gf := &scanFile{
		lc:     0,
		blines: make([]line, x.options.extra.actualBeforeContextLines),
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
		x.cli.stats.IncrScannedLC(int(gf.lc))
	}

	if x.options.Quiet && !x.result.alreadyMatchContent && len(gf.matchedContents) > 0 {
		x.result.alreadyMatchContent = true
	}

	return gf.matchedContents, nil
}

func (x *xfg) isMatchLine(line string) bool {
	if x.options.IgnoreCase && len(x.extra.searchGrepi) > 0 {
		for _, sgr := range x.extra.searchGrepi {
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

	if len(x.extra.searchGrepRe) > 0 {
		for _, re := range x.extra.searchGrepRe {
			if !isMatchRegexp(line, re) {
				return false
			}
		}
	}

	return true // OK, match all
}

func (x *xfg) processContentLine(gf *scanFile) {
	if x.isMatchLine(gf.l) {
		if !x.options.ShowMatchCount && x.options.extra.withBeforeContextLines {
			for _, bl := range gf.blines {
				if bl.lc == 0 {
					continue // skip
				}
				gf.matchedContents = append(gf.matchedContents, bl)
			}
			gf.blines = make([]line, x.options.extra.actualBeforeContextLines)
		}

		if x.options.ShowMatchCount {
			gf.l = ""
		}

		x.optimizeLine(gf)
		gf.matchedContents = append(gf.matchedContents, line{lc: gf.lc, content: gf.l, matched: true})

		if !x.options.ShowMatchCount && x.options.extra.withAfterContextLines {
			gf.aline = x.options.extra.actualAfterContextLines // start countdown for `aline`
		}
	} else {
		if !x.options.ShowMatchCount {
			if x.options.extra.withAfterContextLines && gf.aline > 0 {
				gf.aline--
				x.optimizeLine(gf)
				gf.matchedContents = append(gf.matchedContents, line{lc: gf.lc, content: gf.l})
			} else if x.options.extra.withBeforeContextLines {
				// rotate blines
				// join "2nd to last elements of `blines`" and "current `line`"
				x.optimizeLine(gf)
				gf.blines = append(gf.blines[1:], line{lc: gf.lc, content: gf.l})
			}
		}
	}
}

func (x *xfg) optimizeLine(gf *scanFile) {
	if x.options.MaxColumns > 0 && len(gf.l) > int(x.options.MaxColumns) {
		gf.l = gf.l[:x.options.MaxColumns]
	}
}
