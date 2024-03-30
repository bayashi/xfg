package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	a "github.com/bayashi/actually"
)

// No args, then put help message
func TestArgsNoArgs(t *testing.T) {
	var errOutput bytes.Buffer
	cli := &runner{
		err: &errOutput,
	}

	resetFlag()
	stubExit()
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
	stubExit()
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
	stubExit()
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
		args          []string
		prepareExpect func(o *options)
	}{
		"only path arg": {
			args: []string{"foo"},
			prepareExpect: func(o *options) {
				o.searchPath = []string{"foo"}
			},
		},
		"only specific path arg": {
			args: []string{"--path", "foo"},
			prepareExpect: func(o *options) {
				o.searchPath = []string{"foo"}
			},
		},
		"specific multiple paths": {
			args: []string{"--path", "foo", "--path", "bar"},
			prepareExpect: func(o *options) {
				o.searchPath = []string{"foo", "bar"}
			},
		},
		"path and grep arg": {
			args: []string{"foo", "bar"},
			prepareExpect: func(o *options) {
				o.searchPath = []string{"foo"}
				o.searchGrep = []string{"bar"}
			},
		},
		"path and specific grep args": {
			args: []string{"foo", "--grep", "bar"},
			prepareExpect: func(o *options) {
				o.searchPath = []string{"foo"}
				o.searchGrep = []string{"bar"}
			},
		},
		"path and specific multiple greps": {
			args: []string{"foo", "--grep", "bar", "--grep", "baz"},
			prepareExpect: func(o *options) {
				o.searchPath = []string{"foo"}
				o.searchGrep = []string{"bar", "baz"}
			},
		},
	} {
		t.Run(tname, func(t *testing.T) {
			resetFlag()
			stubExit()
			os.Args = append([]string{fakeCmd}, tt.args...)
			cli := &runner{}

			o := cli.parseArgs()

			expectOptions := defaultOptions()
			tt.prepareExpect(expectOptions)

			a.Got(o).Expect(expectOptions).Same(t)
			a.Got(stubCalled).False(t)
			a.Got(stubCode).Expect(exitOK).Same(t)
		})
	}
}
