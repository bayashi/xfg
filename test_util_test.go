package main

import flag "github.com/spf13/pflag"

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

func defaultOptions() *options {
	return &options{
		searchPath:       []string{},
		searchGrep:       []string{},
		searchStart:      ".",
		groupSeparator:   "--",
		indent:           " ",
		colorPath:        "cyan",
		colorContent:     "red",
		ignore:           []string{},
		ignoreCase:       false,
		relax:            false,
		noColor:          false,
		abs:              false,
		showMatchCount:   false,
		onlyMatch:        false,
		noGroupSeparator: false,
		noIndent:         false,
		hidden:           false,
		skipGitIgnore:    false,
		searchAll:        false,
		contextLines:     0,
		maxMatchCount:    0,
	}
}
