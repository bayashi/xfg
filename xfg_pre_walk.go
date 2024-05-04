package main

import "github.com/bayashi/xfg/internal/xfgutil"

func (x *xfg) preWalkDir() error {
	if !x.options.SkipGitIgnore {
		x.gitignore = prepareGitIgnore(x.cli.homeDir, x.options.SearchStart)
	}

	if !x.options.SkipXfgIgnore {
		x.xfgignore = prepareXfgIgnore(x.cli.homeDir, x.options.XfgIgnoreFile)
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
			x.searchPathRe = searchPathRe
		}
	}

	if len(x.options.SearchGrepRe) > 0 {
		if searchGrepRe, err := xfgutil.CompileRegexps(x.options.SearchGrepRe, !x.options.NotWordBoundary); err != nil {
			return err
		} else {
			x.searchGrepRe = searchGrepRe
		}
	}

	return nil
}

func (x *xfg) prepareRe() error {
	if searchPathi, err := xfgutil.CompileRegexpsIgnoreCase(x.options.SearchPath); err != nil {
		return err
	} else {
		x.searchPathi = searchPathi
	}

	if len(x.options.SearchGrep) > 0 {
		if searchGrepi, err := xfgutil.CompileRegexpsIgnoreCase(x.options.SearchGrep); err != nil {
			return err
		} else {
			x.searchGrepi = searchGrepi
		}
	}

	if len(x.options.Ignore) > 0 {
		if ignoreRe, err := xfgutil.CompileRegexpsIgnoreCase(x.options.Ignore); err != nil {
			return err
		} else {
			x.ignoreRe = ignoreRe
		}
	}

	return nil
}
