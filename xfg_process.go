package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/bayashi/xfg/internal/xfglangxt"
)

func (x *xfg) process() error {
	if err := x.preWalkDir(); err != nil {
		return fmt.Errorf("preWalkDir() : %w", err)
	}

	if x.options.Stats {
		x.cli.stats.Mark("preWalkDir")
	}

	eg := new(errgroup.Group)
	x.walkDir(eg, x.options.SearchStart)
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("walkDir Wait : %w", err)
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
				x.walkDir(eg, filepath.Join(dirPath, s.Name())) // recursively
			}
			s := s
			eg.Go(func() error {
				return x.walkFile(filepath.Join(dirPath, s.Name()), s)
			})
		}

		return nil
	})
}

func (x *xfg) walkFile(fPath string, fInfo fs.DirEntry) error {
	if x.options.Stats {
		x.cli.stats.IncrWalkedPaths()
	}

	if x.isSkippablePath(fPath, fInfo) {
		return nil
	}

	if x.options.Stats {
		x.cli.stats.IncrWalkedContents()
	}

	return x.postMatchPath(fPath, fInfo)
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

func (x *xfg) isMatchFileType(fPath string, fInfo fs.DirEntry) bool {
	switch x.options.Type {
	case "d", "directory":
		return fInfo.IsDir()
	case "l", "symlink":
		return (fInfo.Type() & fs.ModeSymlink) == fs.ModeSymlink
	case "x", "executable":
		if fInfo.IsDir() || (fInfo.Type()&fs.ModeSymlink) == fs.ModeSymlink {
			return false
		}
		i, err := fInfo.Info()
		if err != nil {
			return false // trap error
		}
		return (i.Mode() & 0111) == 0111
	case "e", "empty":
		if fInfo.IsDir() {
			d, err := os.ReadDir(fPath)
			if err != nil {
				return false // trap error
			}
			return len(d) == 0
		} else {
			i, err := fInfo.Info()
			if err != nil {
				return false // trap error
			}
			return i.Size() == 0
		}
	case "s", "socket":
		return (fInfo.Type() & fs.ModeSocket) == fs.ModeSocket
	case "p", "pipe":
		return (fInfo.Type() & fs.ModeNamedPipe) == fs.ModeNamedPipe
	case "b", "block-device":
		return (fInfo.Type() & fs.ModeDevice) == fs.ModeDevice
	case "c", "char-device":
		return (fInfo.Type() & fs.ModeCharDevice) == fs.ModeCharDevice
	default:
		panic("not support type") // unreachable here though
	}
}

func (x *xfg) isSkippablePath(fPath string, fInfo fs.DirEntry) bool {
	if x.options.Quiet && x.hasMatchedAny() {
		return true // already match. skip after all
	}

	if !x.options.SearchAll {
		if (len(x.options.Ext) > 0 && !x.isMatchExt(fInfo, x.options.Ext)) ||
			(len(x.options.Lang) > 0 && !x.isLangFile(fInfo)) ||
			(x.options.Type != "" && !x.isMatchFileType(fPath, fInfo)) {
			return true
		}
	}

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

func (x *xfg) hasMatchedAny() bool {
	x.result.mu.RLock()
	defer x.result.mu.RUnlock()
	if (len(x.options.SearchGrep) == 0 && len(x.result.paths) > 0) ||
		(len(x.options.SearchGrep) > 0 && len(x.result.paths) > 0 && x.result.alreadyMatchContent) {
		return true // already match
	}

	return false
}