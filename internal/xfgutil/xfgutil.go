package xfgutil

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"

	"github.com/mattn/go-isatty"
	"golang.org/x/term"
)

func Procs() int {
	cpu := runtime.NumCPU()
	if cpu == 1 {
		cpu = 2
	}

	runtime.GOMAXPROCS(cpu)

	return cpu
}

func IsTTY() bool {
	fd := os.Stdout.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

func HomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return homeDir, nil
}

func GetTermWindowRows(fd int) (int, error) {
	_, rows, err := term.GetSize(fd)
	if err != nil {
		return 0, err
	}

	return rows, nil
}

func Output(writer *bufio.Writer, out string) error {
	if _, err := fmt.Fprint(writer, out); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}

func CompileRegexps(regexps []string, wordBoundary bool) ([]*regexp.Regexp, error) {
	reList := make([]string, 0, len(regexps))
	for _, re := range regexps {
		if wordBoundary {
			re = "\\b(" + re + ")\\b"
		} else {
			re = "(" + re + ")"
		}
		reList = append(reList, re)
	}

	compiledRegexps := make([]*regexp.Regexp, 0, len(reList))
	for _, re := range reList {
		compiledRe, err := regexp.Compile(re)
		if err != nil {
			return nil, err
		}
		compiledRegexps = append(compiledRegexps, compiledRe)
	}

	return compiledRegexps, nil
}

func CompileRegexpsIgnoreCase(regexps []string) ([]*regexp.Regexp, error) {
	compiledRegexps := make([]*regexp.Regexp, 0, len(regexps))
	for _, re := range regexps {
		compiledRe, err := regexp.Compile("(?i)(" + regexp.QuoteMeta(re) + ")")
		if err != nil {
			return nil, err
		}
		compiledRegexps = append(compiledRegexps, compiledRe)
	}

	return compiledRegexps, nil
}
