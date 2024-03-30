package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	a "github.com/bayashi/actually"
	flag "github.com/spf13/pflag"
)

var (
	stubCalled bool
	stubCode   int
)

func stub() {
	stubCalled = false
	stubCode = 0

	funcExit = func(code int) {
		stubCalled = true
		stubCode = code
	}
}

const fakeCmd = "fake-command"

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

func resetFlag() {
	flag.CommandLine = flag.NewFlagSet(fakeCmd, 1)
}

// No args, then put help message
func TestArgsNoArgs(t *testing.T) {
	var errOutput bytes.Buffer
	cli := &runner{
		err: &errOutput,
	}

	resetFlag()
	stub()
	os.Args = []string{fakeCmd}
	o := cli.parseArgs()

	a.Got(o).Expect(defaultOptions()).Same(t)

	a.Got(stubCalled).True(t)
	a.Got(stubCode).Expect(exitOK).Same(t)

	a.Got(strings.HasPrefix(errOutput.String(), "Version ")).True(t)
	a.Got(errOutput.String()).Expect(`\nUsage: `).Match(t)
	a.Got(errOutput.String()).Expect(`\nOptions:\n`).Match(t)
	a.Got(errOutput.String()).Expect(`-v,\s*--version`).Match(t)
}

func TestArgsHelp(t *testing.T) {
	var errOutput bytes.Buffer
	cli := &runner{
		err: &errOutput,
	}

	resetFlag()
	stub()
	os.Args = []string{fakeCmd, "--help"}
	o := cli.parseArgs()

	a.Got(o).Expect(defaultOptions()).Same(t)

	a.Got(stubCalled).True(t)
	a.Got(stubCode).Expect(exitOK).Same(t)

	a.Got(strings.HasPrefix(errOutput.String(), "Version ")).True(t)
	a.Got(errOutput.String()).Expect(`\nUsage: `).Match(t)
	a.Got(errOutput.String()).Expect(`\nOptions:\n`).Match(t)
	a.Got(errOutput.String()).Expect(`-h,\s*--help`).Match(t)
}

func TestArgsVersion(t *testing.T) {
	var errOutput bytes.Buffer
	cli := &runner{
		err: &errOutput,
	}

	resetFlag()
	stub()
	os.Args = []string{fakeCmd, "--version"}
	o := cli.parseArgs()

	a.Got(o).Expect(defaultOptions()).Same(t)

	a.Got(stubCalled).True(t)
	a.Got(stubCode).Expect(exitOK).Same(t)

	a.Got(strings.HasPrefix(errOutput.String(), "Version ")).True(t)
	a.Got(errOutput.String()).Expect(`\(compiled:`).Match(t)
}

func TestArgs(t *testing.T) {
	for tname, tt := range map[string]struct {
		args   []string
		expect func(o *options)
	}{
		"only path arg": {
			args: []string{"foo"},
			expect: func(o *options) {
				o.searchPath = []string{"foo"}
			},
		},
		"only specific path arg": {
			args: []string{"--path", "foo"},
			expect: func(o *options) {
				o.searchPath = []string{"foo"}
			},
		},
		"specific multiple paths": {
			args: []string{"--path", "foo", "--path", "bar"},
			expect: func(o *options) {
				o.searchPath = []string{"foo", "bar"}
			},
		},
		"path and grep arg": {
			args: []string{"foo", "bar"},
			expect: func(o *options) {
				o.searchPath = []string{"foo"}
				o.searchGrep = []string{"bar"}
			},
		},
		"path and specific grep args": {
			args: []string{"foo", "--grep", "bar"},
			expect: func(o *options) {
				o.searchPath = []string{"foo"}
				o.searchGrep = []string{"bar"}
			},
		},
		"path and specific multiple greps": {
			args: []string{"foo", "--grep", "bar", "--grep", "baz"},
			expect: func(o *options) {
				o.searchPath = []string{"foo"}
				o.searchGrep = []string{"bar", "baz"}
			},
		},
	} {
		t.Run(tname, func(t *testing.T) {
			resetFlag()
			stub()
			os.Args = append([]string{fakeCmd}, tt.args...)
			cli := &runner{}

			o := cli.parseArgs()

			expectOptions := defaultOptions()
			tt.expect(expectOptions)

			a.Got(o).Expect(expectOptions).Same(t)
			a.Got(stubCalled).False(t)
			a.Got(stubCode).Expect(exitOK).Same(t)
		})
	}
}
