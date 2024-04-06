package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	here "github.com/MakeNowJust/heredoc/v2"
	a "github.com/bayashi/actually"
)

const noMatchKeyword = "#NotMatch:4770&4cd-fe9cf87_29706c1@8ab965d!$% ;-P"

func TestMain_OK(t *testing.T) {
	resetFlag()
	stubExit()
	os.Args = []string{fakeCmd, noMatchKeyword, "--no-pager"}
	main()
	a.Got(stubCalled).True(t)
	a.Got(stubCode).Expect(exitOK).Same(t)
}

func TestRun_OK(t *testing.T) {
	var outOutput bytes.Buffer
	cli := &runner{
		out: &outOutput,
	}
	resetFlag()
	stubExit()
	os.Args = []string{fakeCmd, noMatchKeyword, "--no-pager"}
	cli.run()
	a.Got(stubCalled).False(t)
	a.Got(stubCode).Expect(exitOK).Same(t)
	a.Got(outOutput.String()).Expect("").Same(t)
}

func TestXfg_OK(t *testing.T) {
	for tname, tt := range map[string]struct {
		opt            *options
		expect         string
		expectExitCode int
	}{
		"service-b": {
			opt: &options{
				SearchPath: []string{"service-b"},
			},
			expect: here.Doc(`
                testdata/service-b/
                testdata/service-b/main.go
			`),
			expectExitCode: exitOK,
		},
		"service-b grep": {
			opt: &options{
				SearchPath: []string{"service-b"},
				SearchGrep: []string{"func"},
				Indent:     defaultIndent,
			},
			expect: here.Doc(`
                testdata/service-b/
                testdata/service-b/main.go
                 3: func main() {
			`),
			expectExitCode: exitOK,
		},
		"service grep relax": {
			opt: &options{
				SearchPath: []string{"main"},
				SearchGrep: []string{"package b"},
				Indent:     defaultIndent,
				Relax:      true,
			},
			expect: here.Doc(`
                testdata/service-a/main.go
                testdata/service-b/main.go
                 1: package b

                testdata/service-c/main.go
                testdata/service-h/main.go
			`),
			expectExitCode: exitOK,
		},
		"service-b grep bar with C1": {
			opt: &options{
				SearchPath:   []string{"service-b"},
				SearchGrep:   []string{"main"},
				Indent:       defaultIndent,
				ContextLines: 1,
			},
			expect: here.Doc(`
				testdata/service-b/
				testdata/service-b/main.go
				 2: 
				 3: func main() {
				 4: 	bar := 34
			`),
			expectExitCode: exitOK,
		},
		"service-b grep bar with C2": {
			opt: &options{
				SearchPath:   []string{"service-b"},
				SearchGrep:   []string{"main"},
				Indent:       defaultIndent,
				ContextLines: 2,
			},
			expect: here.Doc(`
				testdata/service-b/
				testdata/service-b/main.go
				 1: package b
				 2: 
				 3: func main() {
				 4: 	bar := 34
				 5: }
			`),
			expectExitCode: exitOK,
		},
		"service-c grep 56 with C2. Match 2 consecutive lines": {
			opt: &options{
				SearchPath:   []string{"service-c"},
				SearchGrep:   []string{"56"},
				Indent:       defaultIndent,
				ContextLines: 2,
			},
			expect: here.Doc(`
				testdata/service-c/
				testdata/service-c/main.go
				 2: 
				 3: func main() {
				 4: 	baz := 56
				 5: 	bag := 56
				 6: 
				 7: 	foo()
			`),
			expectExitCode: exitOK,
		},
		"service-a grep foo onlyMatch": {
			opt: &options{
				SearchPath:       []string{"service-a"},
				SearchGrep:       []string{"foo"},
				Indent:           defaultIndent,
				OnlyMatchContent: true,
			},
			expect: here.Doc(`
                testdata/service-a/main.go
                 4: 	foo := 12
			`),
			expectExitCode: exitOK,
		},
		"service-a grep foo": {
			opt: &options{
				SearchPath: []string{"service-a"},
				SearchGrep: []string{"foo"},
				Indent:     defaultIndent,
			},
			expect: here.Doc(`
                testdata/service-a/
                testdata/service-a/a.dat
                testdata/service-a/b
                testdata/service-a/main.go
                 4: 	foo := 12
			`),
			expectExitCode: exitOK,
		},
		"service-c grep foo": {
			opt: &options{
				SearchPath: []string{"service-c"},
				SearchGrep: []string{"foo"},
				Indent:     defaultIndent,
			},
			expect: here.Doc(`
                testdata/service-c/
                testdata/service-c/main.go
                 7: 	foo()
                 10: func foo() {
			`),
			expectExitCode: exitOK,
		},
		"service-c grep foo noIndent": {
			opt: &options{
				SearchPath: []string{"service-c"},
				SearchGrep: []string{"foo"},
				Indent:     defaultIndent,
				NoIndent:   true,
			},
			expect: here.Doc(`
                testdata/service-c/
                testdata/service-c/main.go
                7: 	foo()
                10: func foo() {
			`),
			expectExitCode: exitOK,
		},
		"service-b grep custom indent string": {
			opt: &options{
				SearchPath: []string{"service-b"},
				SearchGrep: []string{"func"},
				Indent:     "	",
			},
			expect: here.Doc(`
                testdata/service-b/
                testdata/service-b/main.go
                	3: func main() {
			`),
			expectExitCode: exitOK,
		},
		"service-d ignore .gitkeep": {
			opt: &options{
				SearchPath: []string{"service-d"},
			},
			expect: here.Doc(`
                testdata/service-d/
			`),
			expectExitCode: exitOK,
		},
		"not pick .gitkeep even with --hidden option": {
			opt: &options{
				SearchPath: []string{"service-d"},
				Hidden:     true,
			},
			expect: here.Doc(`
                testdata/service-d/
			`),
			expectExitCode: exitOK,
		},
		"not pick dotfile by default": {
			opt: &options{
				SearchPath: []string{"service-e"},
				Hidden:     false, // default false
			},
			expect: here.Doc(`
                testdata/service-e/
			`),
			expectExitCode: exitOK,
		},
		"pick dotfile with --hidden option": {
			opt: &options{
				SearchPath: []string{"service-e"},
				Hidden:     true,
			},
			expect: here.Doc(`
                testdata/service-e/
                testdata/service-e/.config
			`),
			expectExitCode: exitOK,
		},
		"not pick up ignorez dir due to .gitignore": {
			opt: &options{
				SearchPath: []string{"service-f"},
			},
			expect: here.Doc(`
                testdata/service-f/
			`),
			expectExitCode: exitOK,
		},
		"pick up ignorez dir with --skip-gitignore option": {
			opt: &options{
				SearchPath:    []string{"service-f"},
				SkipGitIgnore: true,
			},
			expect: here.Doc(`
                testdata/service-f/
                testdata/service-f/ignorez/
			`),
			expectExitCode: exitOK,
		},
		"ignore *min.js by default": {
			opt: &options{
				SearchPath: []string{"service-g"},
			},
			expect: here.Doc(`
                testdata/service-g/
			`),
			expectExitCode: exitOK,
		},
		"ignore option": {
			opt: &options{
				SearchPath: []string{"service-a"},
				Ignore:     []string{"a.dat", "b"},
			},
			expect: here.Doc(`
                testdata/service-a/
                testdata/service-a/main.go
			`),
			expectExitCode: exitOK,
		},
		"ignore anyway even with --hidden option": {
			opt: &options{
				SearchPath: []string{"service-e"},
				Hidden:     true,
				Ignore:     []string{".config"},
			},
			expect: here.Doc(`
                testdata/service-e/
			`),
			expectExitCode: exitOK,
		},
		"ignore anyway even with --search-all option": {
			opt: &options{
				SearchPath: []string{"service-e"},
				SearchAll:  true,
				Ignore:     []string{".config"},
			},
			expect: here.Doc(`
                testdata/service-e/
			`),
			expectExitCode: exitOK,
		},
		"pick *min.js with --search-all option": {
			opt: &options{
				SearchPath: []string{"service-g"},
				SearchAll:  true,
			},
			expect: here.Doc(`
                testdata/service-g/
                testdata/service-g/service-g.min.js
			`),
			expectExitCode: exitOK,
		},
		"pick up ignorez dir with --search-all option": {
			opt: &options{
				SearchPath: []string{"service-f"},
				SearchAll:  true,
			},
			expect: here.Doc(`
                testdata/service-f/
                testdata/service-f/ignorez/
			`),
			expectExitCode: exitOK,
		},
		"service-h": {
			opt: &options{
				SearchPath: []string{"service-h"},
				SearchGrep: []string{"h"},
				Indent:     defaultIndent,
			},
			expect: here.Doc(`
                testdata/service-h/
                testdata/service-h/main.go
                 1: package h
                 4: 	hi()
                 5: 	hello()
                 8: func hi() {
                 11: func hello() {
			`),
			expectExitCode: exitOK,
		},
		"service-h with maxMatchCount": {
			opt: &options{
				SearchPath:    []string{"service-h"},
				SearchGrep:    []string{"h"},
				Indent:        defaultIndent,
				MaxMatchCount: 3,
			},
			expect: here.Doc(`
                testdata/service-h/
                testdata/service-h/main.go
                 1: package h
                 4: 	hi()
                 5: 	hello()
			`),
			expectExitCode: exitOK,
		},
		"service-h show count": {
			opt: &options{
				SearchPath:     []string{"service-h"},
				SearchGrep:     []string{"h"},
				ShowMatchCount: true,
			},
			expect: here.Doc(`
                testdata/service-h/
                testdata/service-h/main.go:5
			`),
			expectExitCode: exitOK,
		},
		"service-c with contextLines": {
			opt: &options{
				SearchPath:     []string{"service-c"},
				SearchGrep:     []string{"func"},
				GroupSeparator: defaultGroupSeparator,
				Indent:         defaultIndent,
				ContextLines:   1,
			},
			expect: here.Doc(`
                testdata/service-c/
                testdata/service-c/main.go
                 2: 
                 3: func main() {
                 4: 	baz := 56
                 --
                 9: 
                 10: func foo() {
                 11: 	println("Result")
			`),
			expectExitCode: exitOK,
		},
		"service-c with contextLines groupSeparator": {
			opt: &options{
				SearchPath:     []string{"service-c"},
				SearchGrep:     []string{"func"},
				GroupSeparator: "====",
				Indent:         defaultIndent,
				ContextLines:   1,
			},
			expect: here.Doc(`
                testdata/service-c/
                testdata/service-c/main.go
                 2: 
                 3: func main() {
                 4: 	baz := 56
                 ====
                 9: 
                 10: func foo() {
                 11: 	println("Result")
			`),
			expectExitCode: exitOK,
		},
		"service-b ignore case to match": {
			opt: &options{
				SearchPath: []string{"Service-B"},
				SearchGrep: []string{"FunC"},
				Indent:     defaultIndent,
				IgnoreCase: true,
			},
			expect: here.Doc(`
                testdata/service-b/
                testdata/service-b/main.go
                 3: func main() {
			`),
			expectExitCode: exitOK,
		},
		"service-c with afterContextLines": {
			opt: &options{
				SearchPath:        []string{"service-c"},
				SearchGrep:        []string{"func"},
				GroupSeparator:    defaultGroupSeparator,
				Indent:            defaultIndent,
				AfterContextLines: 1,
			},
			expect: here.Doc(`
                testdata/service-c/
                testdata/service-c/main.go
                 3: func main() {
                 4: 	baz := 56
                 --
                 10: func foo() {
                 11: 	println("Result")
			`),
			expectExitCode: exitOK,
		},
		"service-c with beforeContextLines": {
			opt: &options{
				SearchPath:         []string{"service-c"},
				SearchGrep:         []string{"func"},
				GroupSeparator:     defaultGroupSeparator,
				Indent:             defaultIndent,
				BeforeContextLines: 1,
			},
			expect: here.Doc(`
                testdata/service-c/
                testdata/service-c/main.go
                 2: 
                 3: func main() {
                 --
                 9: 
                 10: func foo() {
			`),
			expectExitCode: exitOK,
		},
		"service-b quiet": {
			opt: &options{
				SearchPath: []string{"service-b"},
				Quiet:      true,
			},
			expect:         "",
			expectExitCode: exitOK,
		},
		"service-b quiet no match": {
			opt: &options{
				SearchPath: []string{noMatchKeyword},
				Quiet:      true,
			},
			expect:         "",
			expectExitCode: exitErr,
		},
		"service-b quiet no match both search and grep": {
			opt: &options{
				SearchPath: []string{noMatchKeyword},
				SearchGrep: []string{noMatchKeyword},
				Quiet:      true,
			},
			expect:         "",
			expectExitCode: exitErr,
		},
		"service-b grep bar with C1 and max-columns": {
			opt: &options{
				SearchPath:   []string{"service-b"},
				SearchGrep:   []string{"main"},
				Indent:       defaultIndent,
				ContextLines: 1,
				MaxColumns:   7,
			},
			expect: here.Doc(`
				testdata/service-b/
				testdata/service-b/main.go
				 2: 
				 3: func ma
				 4: 	bar :=
			`),
			expectExitCode: exitOK,
		},
		"not pick up ignorex dir due to .xfgignore": {
			opt: &options{
				SearchPath:    []string{"service-i"},
				XfgIgnoreFile: filepath.Join("testdata", ".xfgignore"),
			},
			expect: here.Doc(`
                testdata/service-i/
			`),
			expectExitCode: exitOK,
		},
		"pick up ignorex dir with --skip-xfgignore option": {
			opt: &options{
				SearchPath:    []string{"service-i"},
				XfgIgnoreFile: filepath.Join("testdata", ".xfgignore"),
				SkipXfgIgnore: true,
			},
			expect: here.Doc(`
                testdata/service-i/
                testdata/service-i/ignorex/
			`),
			expectExitCode: exitOK,
		},
		"pick up ignorex dir with --search-all option": {
			opt: &options{
				SearchPath:    []string{"service-i"},
				XfgIgnoreFile: filepath.Join("testdata", ".xfgignore"),
				SearchAll:     true,
			},
			expect: here.Doc(`
                testdata/service-i/
                testdata/service-i/ignorex/
			`),
			expectExitCode: exitOK,
		},
		"service-c with contextLines, but no-group-separator": {
			opt: &options{
				SearchPath:       []string{"service-c"},
				SearchGrep:       []string{"func"},
				GroupSeparator:   defaultGroupSeparator,
				Indent:           defaultIndent,
				ContextLines:     1,
				NoGroupSeparator: true,
			},
			expect: here.Doc(`
                testdata/service-c/
                testdata/service-c/main.go
                 2: 
                 3: func main() {
                 4: 	baz := 56
                 9: 
                 10: func foo() {
                 11: 	println("Result")
			`),
			expectExitCode: exitOK,
		},
	} {
		t.Run(tname, func(t *testing.T) {
			var o bytes.Buffer
			cli := &runner{
				out:   &o,
				isTTY: true,
			}

			tt.opt.NoPager = true
			tt.opt.NoColor = true
			tt.opt.SearchStart = "./testdata"

			code, err := cli.xfg(tt.opt)

			a.Got(err).Debug("options", tt.opt).NoError(t)
			a.Got(code).Expect(tt.expectExitCode).Same(t)

			tt.expect = windowsBK(tt.expect)
			a.Got(o.String()).Expect(tt.expect).X().Debug("options", tt.opt).Same(t)
		})
	}
}

func TestNonTTY(t *testing.T) {
	resetFlag()
	stubExit()
	os.Args = []string{fakeCmd, "service-b", "func", "-s", "./testdata"}
	var o bytes.Buffer
	cli := &runner{
		out:   &o,
		isTTY: false,
	}

	cli.run()

	// no color, no pager
	expect := here.Doc(`
	    testdata/service-b/main.go:3:func main() {
	`)

	expect = windowsBK(expect)

	a.Got(o.String()).Expect(expect).X().Same(t)
}
