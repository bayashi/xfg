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
	ignore "github.com/sabhiram/go-gitignore"
)

func defaultOptions() *options {
	return &options{
		SearchStart:    ".",
		Indent:         defaultIndent,
		GroupSeparator: defaultGroupSeparator,
		ColorPathBase:  "yellow",
		ColorPath:      "cyan",
		ColorContent:   "red",
	}
}

const XFG_RC_ENV_KEY = "XFG_RC_FILE_PATH"

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

func prepareGitIgnore(homeDir string, sPath string) *ignore.GitIgnore {
	const GIT_IGNOE_FILE_NAME = ".gitignore"
	// read .gitignore file in start directory to search or home directory
	// There would be no .gitignore file, then `gitignore` variable will be `nil`.
	gitignore, _ := ignore.CompileIgnoreFile(filepath.Join(sPath, GIT_IGNOE_FILE_NAME))
	if gitignore == nil {
		gitignore, _ = ignore.CompileIgnoreFile(filepath.Join(homeDir, GIT_IGNOE_FILE_NAME))
	}

	return gitignore
}

func prepareXfgIgnore(homeDir string, xfgignoreFilePath string) *ignore.GitIgnore {
	if xfgignoreFilePath != "" {
		xfgignore, _ := ignore.CompileIgnoreFile(xfgignoreFilePath)
		if xfgignore != nil {
			return xfgignore
		}
	}

	const XFG_IGNOE_FILE_NAME = ".xfgignore"
	// read .xfgignore file in XDG Base directory or home directory
	// There would be no .xfgignore file, then `xfgignore` variable will be `nil`.
	xfgignore, _ := ignore.CompileIgnoreFile(filepath.Join(xdg.ConfigHome, XFG_IGNOE_FILE_NAME))
	if xfgignore == nil {
		xfgignore, _ = ignore.CompileIgnoreFile(filepath.Join(homeDir, XFG_IGNOE_FILE_NAME))
	}

	return xfgignore
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

func validateStartPath(startPath string) error {
	d, err := os.Stat(startPath)
	if err != nil {
		return fmt.Errorf("wrong path `%s` : %w", startPath, err)
	}

	if !d.IsDir() {
		return fmt.Errorf("path `%s` should point to a directory", startPath)
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

func canSkipStuff(fInfo fs.DirEntry) bool {
	return !fInfo.IsDir() && (fInfo.Name() == ".gitkeep" || strings.HasSuffix(fInfo.Name(), ".min.js"))
}
