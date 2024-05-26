package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
	"github.com/bayashi/xfg/internal/xfglangxt"
)

func defaultOptions() *options {
	return &options{
		SearchStart:    []string{"."},
		Indent:         defaultIndent,
		GroupSeparator: defaultGroupSeparator,
		ColorPathBase:  "yellow",
		ColorPath:      "cyan",
		ColorContent:   "red",
		MaxDepth:       255,
	}
}

func readRC(homeDir string) (*options, error) {
	o := defaultOptions()

	if xfgRCFilePath := os.Getenv(XFG_RC_ENV_KEY); xfgRCFilePath != "" {
		if _, err := toml.DecodeFile(xfgRCFilePath, &o); err != nil {
			return nil, fmt.Errorf("could not decode toml config env:%s `%s` : %w", XFG_RC_ENV_KEY, xfgRCFilePath, err)
		}
		return o, nil
	}

	xfgRCFilePath := filepath.Join(xdg.ConfigHome, XFG_RC_FILE)
	_, err := toml.DecodeFile(xfgRCFilePath, &o)
	if err == nil {
		return o, nil
	} else if !errors.Is(err, syscall.ENOENT) {
		return nil, fmt.Errorf("%s : %w", xfgRCFilePath, err)
	}

	xfgRCFilePath = filepath.Join(homeDir, XFG_RC_FILE)
	_, err = toml.DecodeFile(xfgRCFilePath, &o)
	if err != nil && !errors.Is(err, syscall.ENOENT) {
		return nil, fmt.Errorf("%s : %w", xfgRCFilePath, err)
	}

	return o, nil
}

func isBinaryFile(fh *os.File) (bool, error) {
	dat := make([]byte, 8000)
	n, err := fh.Read(dat)
	if err != nil {
		return false, fmt.Errorf("could not read fh : %w", err)
	}

	for _, c := range dat[:n] {
		if c == 0x00 {
			return true, nil
		}
	}

	return false, nil
}

func isRegularFile(fInfo fs.DirEntry) bool {
	if fInfo.Type() != 0 {
		return false
	}

	fi, err := fInfo.Info()
	return err == nil && fi.Size() > 0
}

func validateStartPath(startPaths []string) error {
	for i, sp := range startPaths {
		sp = filepath.Clean(sp)
		d, err := os.Stat(sp)
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return fmt.Errorf("path `%s` should point to a directory", startPaths[i])
		}

		startPaths[i] = sp
	}

	return nil
}

func validateLanguageCondition(lang []string) error {
	for _, l := range lang {
		if !xfglangxt.IsSupported(l) {
			return fmt.Errorf("`%s` is not supported. --type-list prints all support languages", l)
		}
	}

	return nil
}

func isMatch(target string, included string) bool {
	if target == "" || included == "" {
		return false
	}

	return strings.Contains(target, included)
}

func isMatchRegexp(target string, re *regexp.Regexp) bool {
	if target == "" || re == nil {
		return false
	}

	return re.MatchString(target)
}

func isDefaultSkipFile(fInfo fs.DirEntry) bool {
	return !fInfo.IsDir() && (fInfo.Name() == ".gitkeep" ||
		strings.HasSuffix(fInfo.Name(), ".min.js") || strings.HasSuffix(fInfo.Name(), ".min.css"))
}

func isDefaultSkipDir(fInfo fs.DirEntry) bool {
	return fInfo.IsDir() && (fInfo.Name() == ".git" || fInfo.Name() == ".svn" ||
		fInfo.Name() == "node_modules" ||
		fInfo.Name() == "vendor")
}
