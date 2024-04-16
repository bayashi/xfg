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
	o := cli.parseArgs(&options{})

	a.Got(o).Expect(&options{}).Same(t)

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
	o := cli.parseArgs(defaultOptions())

	e := expectedDefaultOptions()
	a.Got(o).Expect(&e).Same(t)

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
	o := cli.parseArgs(defaultOptions())

	e := expectedDefaultOptions()
	a.Got(o).Expect(&e).Same(t)

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
				o.SearchPath = []string{"foo"}
			},
		},
		"only specific path arg": {
			args: []string{"--path", "foo"},
			prepareExpect: func(o *options) {
				o.SearchPath = []string{"foo"}
			},
		},
		"only specific grep arg": {
			args: []string{"--grep", "foo"},
			prepareExpect: func(o *options) {
				o.SearchGrep = []string{"foo"}
				o.onlyMatchContent = true
			},
		},
		"specific multiple paths": {
			args: []string{"--path", "foo", "--path", "bar"},
			prepareExpect: func(o *options) {
				o.SearchPath = []string{"foo", "bar"}
			},
		},
		"path and grep arg": {
			args: []string{"foo", "bar"},
			prepareExpect: func(o *options) {
				o.SearchPath = []string{"foo"}
				o.SearchGrep = []string{"bar"}
				o.onlyMatchContent = true
			},
		},
		"path and specific grep args": {
			args: []string{"foo", "--grep", "bar"},
			prepareExpect: func(o *options) {
				o.SearchPath = []string{"foo"}
				o.SearchGrep = []string{"bar"}
				o.onlyMatchContent = true
			},
		},
		"path and specific multiple greps": {
			args: []string{"foo", "--grep", "bar", "--grep", "baz"},
			prepareExpect: func(o *options) {
				o.SearchPath = []string{"foo"}
				o.SearchGrep = []string{"bar", "baz"}
				o.onlyMatchContent = true
			},
		},
		"path and multiple grep args": {
			args: []string{"foo", "bar", "baz"},
			prepareExpect: func(o *options) {
				o.SearchPath = []string{"foo"}
				o.SearchGrep = []string{"bar", "baz"}
				o.onlyMatchContent = true
			},
		},
	} {
		t.Run(tname, func(t *testing.T) {
			resetFlag()
			stubExit()
			os.Args = append([]string{fakeCmd}, tt.args...)
			cli := &runner{}
			o := cli.parseArgs(defaultOptions())

			expectOptions := expectedDefaultOptions()
			tt.prepareExpect(&expectOptions)

			a.Got(o).Expect(&expectOptions).Same(t)
			a.Got(stubCalled).False(t)
			a.Got(stubCode).Expect(exitOK).Same(t)
		})
	}
}

func TestPrepareAliases(t *testing.T) {
	o := &options{
		Unrestricted: true,
	}
	o.prepareAliases()
	a.Got(o.SearchAll).True(t)
}
