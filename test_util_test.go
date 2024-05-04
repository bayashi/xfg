package main

import (
	"os"
	"strings"

	flag "github.com/spf13/pflag"
)

const fakeCmd = "fake-command"

var (
	stubCalled bool
	stubCode   int
)

func stubExit() {
	stubCalled = false
	stubCode = 0

	funcExit = func(code int) {
		stubCalled = true
		stubCode = code
	}
}

func resetFlag() {
	flag.CommandLine = flag.NewFlagSet(fakeCmd, 1)
}

func expectedDefaultOptions() options {
	return options{
		SearchPath:       nil,
		SearchGrep:       nil,
		SearchStart:      ".",
		GroupSeparator:   "--",
		Indent:           " ",
		ColorPathBase:    "yellow",
		ColorPath:        "cyan",
		ColorContent:     "red",
		Ignore:           nil,
		IgnoreCase:       false,
		NoColor:          false,
		Abs:              false,
		ShowMatchCount:   false,
		onlyMatchContent: false,
		NoGroupSeparator: false,
		NoIndent:         false,
		Hidden:           false,
		SkipGitIgnore:    false,
		SearchAll:        false,
		flagLangList:     false,
		ContextLines:     0,
		MaxMatchCount:    0,
	}
}

func isWindowsTestRunner() bool {
	return os.Getenv("RUNNER_OS") == "Windows"
}

func windowsBK(src string) string {
	if isWindowsTestRunner() {
		// BK: override path delimiter for Windows
		src = strings.ReplaceAll(src, "/", "\\")
	}

	return src
}
