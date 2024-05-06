package main

import (
	"github.com/bayashi/xfg/internal/xfgignore"
	"github.com/bayashi/xfg/internal/xfgutil"
)

func (x *xfg) preWalkDir() error {
	if !x.options.SkipGitIgnore {
		if ms := xfgignore.SetUpGlobalGitIgnores(x.options.SearchStart, x.cli.homeDir); len(ms) > 0 {
			x.extra.ignoreMatchers = append(x.extra.ignoreMatchers, ms...)
		}
	}

	if !x.options.SkipXfgIgnore {
		if ms := xfgignore.SetupGlobalXFGIgnore(x.options.SearchStart, x.cli.homeDir, x.options.XfgIgnoreFile); len(ms) > 0 {
			x.extra.ignoreMatchers = append(x.extra.ignoreMatchers, ms...)
		}
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
			x.extra.searchPathRe = searchPathRe
		}
	}

	if len(x.options.SearchGrepRe) > 0 {
		if searchGrepRe, err := xfgutil.CompileRegexps(x.options.SearchGrepRe, !x.options.NotWordBoundary); err != nil {
			return err
		} else {
			x.extra.searchGrepRe = searchGrepRe
		}
	}

	return nil
}

func (x *xfg) prepareRe() error {
	if searchPathi, err := xfgutil.CompileRegexpsIgnoreCase(x.options.SearchPath); err != nil {
		return err
	} else {
		x.extra.searchPathi = searchPathi
	}

	if len(x.options.SearchGrep) > 0 {
		if searchGrepi, err := xfgutil.CompileRegexpsIgnoreCase(x.options.SearchGrep); err != nil {
			return err
		} else {
			x.extra.searchGrepi = searchGrepi
		}
	}

	if len(x.options.Ignore) > 0 {
		if ignoreRe, err := xfgutil.CompileRegexpsIgnoreCase(x.options.Ignore); err != nil {
			return err
		} else {
			x.extra.ignoreRe = ignoreRe
		}
	}

	return nil
}
