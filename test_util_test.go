package main

import (
	"runtime"
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

func isWindowsTestRunner() bool {
	return runtime.GOOS == "windows"
}

func windowsBK(src string) string {
	if isWindowsTestRunner() {
		// BK: override path delimiter for Windows
		src = strings.ReplaceAll(src, "/", "\\")
	}

	return src
}
