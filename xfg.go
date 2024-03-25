package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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

type walkerArg struct {
	path      string
	info      fs.FileInfo
	gitignore *ignore.GitIgnore
}

func (x *xfg) Search() error {
	if err := validateStartPath(x.options.searchStart); err != nil {
		return err
	}

	gitignore := x.compileGitIgnore(x.options.searchStart)

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

	if x.isIgnore(fPath) {
		return nil // skip by --ignore option
	}

	if !x.options.searchAll {
		if fInfo.IsDir() && fInfo.Name() == ".git" {
			return filepath.SkipDir // not search for .git directory
		}
	}

	if x.isSkip(fPath, fInfo, wa.gitignore) {
		return nil // skip
	}

	x.onMatchPath(fPath, fInfo)

	return nil
}

func (x *xfg) isIgnore(fPath string) bool {
	for _, i := range x.options.ignore {
		if i != "" && strings.Contains(fPath, i) {
			return true // skip
		}
	}

	return false
}

func (x *xfg) isSkip(fPath string, fInfo fs.FileInfo, gitignore *ignore.GitIgnore) bool {
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

	if !strings.Contains(fPath, x.options.searchPath) {
		return true // not match
	}

	return false
}

func (x *xfg) onMatchPath(fPath string, fInfo fs.FileInfo) (err error) {
	matchedPath := path{
		info: fInfo,
	}

	if x.options.searchGrep != "" && isRegularFile(fInfo) {
		matchedPath.contents, err = x.grepPath(fPath)
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
		matchedPath.path = strings.ReplaceAll(fPath, x.options.searchPath, x.pathHighlighter)
	}

	if fInfo.IsDir() {
		matchedPath.path = matchedPath.path + string(filepath.Separator)
	}

	x.result = append(x.result, matchedPath)

	return nil
}

func (x *xfg) compileGitIgnore(sPath string) *ignore.GitIgnore {
	const GIT_IGNOE_FILE_NAME = ".gitignore"
	var gitignore *ignore.GitIgnore
	if !x.options.skipGitIgnore {
		// read .gitignore file in start directory to search or home directory
		// There would be no .gitignore file, then `gitignore` variable will be `nil`.
		gitignore, _ = ignore.CompileIgnoreFile(filepath.Join(sPath, GIT_IGNOE_FILE_NAME))
		if gitignore == nil {
			if homeDir, err := os.UserHomeDir(); err == nil {
				gitignore, _ = ignore.CompileIgnoreFile(filepath.Join(homeDir, GIT_IGNOE_FILE_NAME))
			}
		}
	}

	return gitignore
}

func (x *xfg) grepPath(fPath string) ([]line, error) {
	fh, err := os.Open(fPath)
	if err != nil {
		return nil, fmt.Errorf("could not open file `%s`: %w", fPath, err)
	}
	defer fh.Close()

	isBinary, err := x.isBinary(fh)
	if err != nil {
		return nil, fmt.Errorf("error during isBinary file `%s`: %w", fPath, err)
	}
	if isBinary {
		return nil, nil
	}

	if _, err := fh.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("could not seek `%s`: %w", fPath, err)
	}

	matchedContents, err := x.grepFile(bufio.NewScanner(fh), fPath)
	if err != nil {
		return nil, fmt.Errorf("could not grepFile `%s`: %w", fPath, err)
	}

	return matchedContents, nil
}

func (x *xfg) isBinary(fh *os.File) (bool, error) {
	dat := make([]byte, 8000)
	n, err := fh.Read(dat)
	if err != nil {
		return false, fmt.Errorf("could not read fh: %w", err)
	}

	for _, c := range dat[:n] {
		if c == 0x00 {
			return true, nil
		}
	}

	return false, nil
}

type grepFile struct {
	lc               int32  // line count
	l                string // line text
	blines           []line // slice for before lines
	aline            uint32 // the count for after lines
	withContextLines bool

	matchedContents []line // result
}

func (x *xfg) grepFile(scanner *bufio.Scanner, fPath string) ([]line, error) {
	gf := &grepFile{
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

func (x *xfg) processContentLine(gf *grepFile) {
	if strings.Contains(gf.l, x.options.searchGrep) {
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
			gf.l = strings.ReplaceAll(gf.l, x.options.searchGrep, x.grepHighlighter)
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

func isRegularFile(fInfo os.FileInfo) bool {
	return fInfo.Size() > 0 && fInfo.Mode().Type() == 0
}

func validateStartPath(startPath string) error {
	d, err := os.Stat(startPath)
	if err != nil {
		return fmt.Errorf("path `%s` is wrong: %w", startPath, err)
	}

	if !d.IsDir() {
		return fmt.Errorf("path `%s` should point to a directory", startPath)
	}

	return nil
}

func output(writer *bufio.Writer, out string) error {
	if _, err := fmt.Fprint(writer, out); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}

func (x *xfg) Show(w io.Writer) error {
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
				x.showContent(&out, p.contents)
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

func (x *xfg) showContent(out *string, contents []line) error {
	var blc int32 = 0
	for _, line := range contents {
		if blc != 0 && line.lc-blc > 1 {
			*out = *out + x.options.indent + x.options.groupSeparator + "\n"
		}
		lc := fmt.Sprintf("%d", line.lc)
		if !x.options.noColor && line.matched {
			lc = x.grepHighlitColor.Sprint(lc)
		}
		*out = *out + fmt.Sprintf("%s%s: %s\n", x.options.indent, lc, line.content)
		blc = line.lc
	}

	return nil
}
