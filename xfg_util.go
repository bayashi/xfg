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
	"golang.org/x/term"
)

func defaultOptions() *options {
	return &options{
		SearchStart:    ".",
		Indent:         defaultIndent,
		GroupSeparator: defaultGroupSeparator,
		ColorPath:      "cyan",
		ColorContent:   "red",
	}
}

const XFG_RC_ENV_KEY = "XFG_RC_FILE_PATH"

func readRC() (*options, error) {
	o := defaultOptions()

	if xfgRCFilePath := os.Getenv(XFG_RC_ENV_KEY); xfgRCFilePath != "" {
		if _, err := toml.DecodeFile(xfgRCFilePath, &o); err != nil {
			return nil, fmt.Errorf("config path from env `%s` : %w", XFG_RC_ENV_KEY, err)
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

	if homeDir, err := os.UserHomeDir(); err == nil {
		xfgRCFilePath := filepath.Join(homeDir, XFG_RC_FILE)
		_, err := toml.DecodeFile(xfgRCFilePath, &o)
		if err != nil && !errors.Is(err, syscall.ENOENT) {
			return nil, fmt.Errorf("%s : %w", xfgRCFilePath, err)
		}
	}

	return o, nil
}

func prepareGitIgnore(sPath string) *ignore.GitIgnore {
	const GIT_IGNOE_FILE_NAME = ".gitignore"
	// read .gitignore file in start directory to search or home directory
	// There would be no .gitignore file, then `gitignore` variable will be `nil`.
	gitignore, _ := ignore.CompileIgnoreFile(filepath.Join(sPath, GIT_IGNOE_FILE_NAME))
	if gitignore == nil {
		if homeDir, err := os.UserHomeDir(); err == nil {
			gitignore, _ = ignore.CompileIgnoreFile(filepath.Join(homeDir, GIT_IGNOE_FILE_NAME))
		}
	}

	return gitignore
}

func prepareXfgIgnore(xfgFilePath string) *ignore.GitIgnore {
	xfgignore, _ := ignore.CompileIgnoreFile(xfgFilePath)
	if xfgignore != nil {
		return xfgignore
	}

	const XFG_IGNOE_FILE_NAME = ".xfgignore"
	// read .xfgignore file in XDG Base directory or home directory
	// There would be no .xfgignore file, then `xfgignore` variable will be `nil`.
	xfgignore, _ = ignore.CompileIgnoreFile(filepath.Join(xdg.ConfigHome, XFG_IGNOE_FILE_NAME))
	if xfgignore == nil {
		if homeDir, err := os.UserHomeDir(); err == nil {
			xfgignore, _ = ignore.CompileIgnoreFile(filepath.Join(homeDir, XFG_IGNOE_FILE_NAME))
		}
	}

	return xfgignore
}

func isBinaryFile(fh *os.File) (bool, error) {
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

func getTermWindowRows(fd int) (int, error) {
	_, rows, err := term.GetSize(fd)
	if err != nil {
		return 0, err
	}

	return rows, nil
}

func canSkipStuff(fInfo fs.FileInfo) bool {
	return !fInfo.IsDir() && (fInfo.Name() == ".gitkeep" || strings.HasSuffix(fInfo.Name(), ".min.js"))
}
