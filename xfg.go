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
	lc      int
	content string
}

type path struct {
	path    string
	info    os.FileInfo
	content []line
}

type xfg struct {
	options *options

	pathHighlightColor *color.Color
	grepHighlightColor *color.Color

	result []path
}

func NewX(o *options, pathHighlightColor *color.Color, grepHighlightColor *color.Color) *xfg {
	x := &xfg{
		options: o,
	}
	if pathHighlightColor != nil {
		x.pathHighlightColor = pathHighlightColor
	}
	if grepHighlightColor != nil {
		x.grepHighlightColor = grepHighlightColor
	}

	return x
}

func (x *xfg) Show(w io.Writer) error {
	for _, p := range x.result {
		if _, err := fmt.Fprintf(w, "%s\n", p.path); err != nil {
			return err
		}
		for _, lines := range p.content {
			if _, err := fmt.Fprintf(w, "  %d: %s\n", lines.lc, lines.content); err != nil {
				return err
			}
		}
		if x.options.relax && len(p.content) > 0 {
			if _, err := fmt.Fprint(w, "\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

func (x *xfg) Search() error {
	sPath, err := validateStartPath(x.options.searchStart)
	if err != nil {
		return err
	}

	hPath := x.pathHighlightColor.Sprintf(x.options.searchPath)
	hGrep := x.grepHighlightColor.Sprintf(x.options.searchGrep)

	var paths []path
	walkErr := filepath.Walk(sPath, func(fPath string, fInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("something went wrong within path `%s` at `%s`: %w", sPath, fPath, err)
		}

		if fInfo.IsDir() && fInfo.Name() == ".git" {
			return filepath.SkipDir
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
			matchedPath.path = strings.ReplaceAll(fPath, x.options.searchPath, hPath)
		}

		if !fInfo.IsDir() && x.options.searchGrep != "" {
			matchedContents, err := matchedContents(fPath, x.options.searchGrep)
			if err != nil {
				return err
			}

			if x.options.noColor {
				matchedPath.content = matchedContents
			} else {
				matchedPath.content = colorMatchedContents(matchedContents, x.options.searchGrep, hGrep)
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

func colorMatchedContents(matchedContents []line, old string, new string) []line {
	var newContents []line
	for _, mc := range matchedContents {
		mc.content = strings.ReplaceAll(mc.content, old, new)
		newContents = append(newContents, mc)
	}

	return newContents
}

func matchedContents(fPath string, g string) ([]line, error) {
	fh, err := os.Open(fPath)
	if err != nil {
		return nil, fmt.Errorf("could not open file `%s`: %w", fPath, err)
	}
	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	var matchedContents []line
	lc := 0
	scanner.Err()
	for scanner.Scan() {
		l := scanner.Text()
		lc++
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("could not scan file `%s` line %d: %w", fPath, lc, err)
		}
		if !strings.Contains(l, g) {
			continue
		}
		matchedContents = append(matchedContents, line{lc: lc, content: l})
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
