package main

import (
	"github.com/bayashi/xfg/internal/xfgignore"
	"github.com/bayashi/xfg/internal/xfgutil"
)

func (x *xfg) preWalkDir() error {
	if x.options.IgnoreCase {
		if err := x.prepareIgnoreCaseRe(); err != nil {
			return err
		}
		if err := x.prepareIgnoreOption(); err != nil {
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

func (x *xfg) prepareIgnoreCaseRe() error {
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

	return nil
}

func (x *xfg) prepareIgnoreOption() error {
	if len(x.options.Ignore) > 0 {
		if ignoreOptionRe, err := xfgutil.CompileRegexpsIgnoreCase(x.options.Ignore); err != nil {
			return err
		} else {
			x.extra.ignoreOptionRe = ignoreOptionRe
		}
	}

	return nil
}

func (x *xfg) initIgnoreMatchers(rootDir string) xfgignore.Matchers {
	var gms xfgignore.Matchers
	if !x.options.SkipGitIgnore {
		if ms := xfgignore.SetUpGlobalGitIgnores(rootDir, x.cli.homeDir); len(ms) > 0 {
			gms = append(gms, ms...)
		}
	}

	if !x.options.SkipXfgIgnore {
		if ms := xfgignore.SetupGlobalXFGIgnore(rootDir, x.cli.homeDir, x.options.XfgIgnoreFile); len(ms) > 0 {
			gms = append(gms, ms...)
		}
	}

	return gms
}
