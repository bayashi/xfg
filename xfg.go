package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

type line struct {
	lc      int32
	content string
	matched bool
}

type path struct {
	path    string
	info    os.FileInfo
	content []line
}

type xfg struct {
	options *options

	pathHighlitColor *color.Color
	pathHighlighter  string
	grepHighlitColor *color.Color
	grepHighlighter  string

	result []path
}

func NewX(o *options, pathHighlightColor *color.Color, grepHighlightColor *color.Color) *xfg {
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

func (x *xfg) Show(w io.Writer) error {
	writer := bufio.NewWriter(w)
	for _, p := range x.result {
		if _, err := fmt.Fprintf(writer, "%s\n", p.path); err != nil {
			return err
		}
		for _, line := range p.content {
			lc := fmt.Sprintf("%d", line.lc)
			if !x.options.noColor && line.matched {
				lc = x.grepHighlitColor.Sprint(lc)
			}
			if _, err := fmt.Fprintf(writer, "  %s: %s\n", lc, line.content); err != nil {
				return err
			}
		}
		if x.options.relax && len(p.content) > 0 {
			if _, err := fmt.Fprint(writer, "\n"); err != nil {
				return err
			}
		}
		if err := writer.Flush(); err != nil {
			return err
		}
	}

	return nil
}

func (x *xfg) Search() error {
	sPath, err := validateStartPath(x.options.searchStart)
	if err != nil {
		return err
	}

	var paths []path
	walkErr := filepath.Walk(sPath, func(fPath string, fInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("something went wrong within path `%s` at `%s`: %w", sPath, fPath, err)
		}

		if fInfo.IsDir() && fInfo.Name() == ".git" {
			return filepath.SkipDir
		}

		if fInfo.IsDir() {
			fPath = fPath + string(filepath.Separator)
		}

		if !strings.Contains(fPath, x.options.searchPath) {
			return nil
		}

		matchedPath := path{
			info: fInfo,
		}

		if x.options.abs {
			absPath, err := filepath.Abs(fPath)
			if err != nil {
				return err
			}
			fPath = absPath
		}

		if x.options.noColor {
			matchedPath.path = fPath
		} else {
			matchedPath.path = strings.ReplaceAll(fPath, x.options.searchPath, x.pathHighlighter)
		}

		if !fInfo.IsDir() && x.options.searchGrep != "" && fInfo.Size() > 0 {
			matchedPath.content, err = x.grep(fPath)
			if err != nil {
				return err
			}
		}

		paths = append(paths, matchedPath)

		return nil
	})
	if walkErr != nil {
		return walkErr
	}

	x.result = paths

	return nil
}

func (x *xfg) grep(fPath string) ([]line, error) {
	fh, err := os.Open(fPath)
	if err != nil {
		return nil, fmt.Errorf("could not open file `%s`: %w", fPath, err)
	}
	defer fh.Close()

	isBinary, err := x.isBinary(fh)
	if err != nil {
		return nil, err
	}
	if isBinary {
		return nil, nil
	}

	fh.Seek(0, 0)

	matchedContents, err := x.grepFile(bufio.NewScanner(fh), fPath)
	if err != nil {
		return nil, err
	}

	return matchedContents, nil
}

func (x *xfg) isBinary(fh *os.File) (bool, error) {
	dat := make([]byte, 8000)
	n, err := fh.Read(dat)
	if err != nil {
		return false, err
	}

	for _, c := range dat[:n] {
		if c == 0x00 {
			return true, nil
		}
	}

	return false, nil
}

func (x *xfg) grepFile(scanner *bufio.Scanner, fPath string) ([]line, error) {
	var (
		lc              int32 = 0 // line count
		matchedContents []line

		blines = make([]line, x.options.contextLines) // slice for before lines
		aline  uint32                                 // the count for after lines

		optC = x.options.contextLines > 0
	)

	for scanner.Scan() {
		lc++
		l := scanner.Text()
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("could not scan file `%s` line %d: %w", fPath, lc, err)
		}
		if strings.Contains(l, x.options.searchGrep) {
			if optC {
				for _, bl := range blines {
					if bl.lc == 0 {
						continue // skip
					}
					matchedContents = append(matchedContents, bl)
				}
				blines = make([]line, x.options.contextLines)
			}

			if !x.options.noColor {
				l = strings.ReplaceAll(l, x.options.searchGrep, x.grepHighlighter)
			}

			matchedContents = append(matchedContents, line{lc: lc, content: l, matched: true})

			if optC {
				aline = x.options.contextLines // start countdown for `aline`
			}
		} else {
			if optC {
				if aline > 0 {
					aline--
					matchedContents = append(matchedContents, line{lc: lc, content: l})
				} else {
					// lotate blines
					// join "2nd to last elements of `blines`" and "current `line`"
					blines = append(blines[1:], line{lc: lc, content: l})
				}
			}
		}
	}

	return matchedContents, nil
}

func validateStartPath(startPath string) (string, error) {
	d, err := os.Stat(startPath)
	if err != nil {
		return "", fmt.Errorf("path `%s` is wrong: %w", startPath, err)
	}

	if !d.IsDir() {
		return "", fmt.Errorf("path `%s` should point to a directory", startPath)
	}

	return startPath, nil
}
