package main

import (
	"bytes"
	"os"
	"strings"
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
		opt    *options
		expect string
	}{
		"service-b": {
			opt: &options{
				searchPath: []string{"service-b"},
			},
			expect: here.Doc(`
                testdata/service-b/
                testdata/service-b/main.go
			`),
		},
		"service-b grep": {
			opt: &options{
				searchPath: []string{"service-b"},
				searchGrep: []string{"func"},
				indent:     defaultIndent,
			},
			expect: here.Doc(`
                testdata/service-b/
                testdata/service-b/main.go
                 3: func main() {
			`),
		},
		"service grep relax": {
			opt: &options{
				searchPath: []string{"main"},
				searchGrep: []string{"package b"},
				indent:     defaultIndent,
				relax:      true,
			},
			expect: here.Doc(`
                testdata/service-a/main.go
                testdata/service-b/main.go
                 1: package b

                testdata/service-c/main.go
                testdata/service-h/main.go
			`),
		},
		"service-b grep bar with C1": {
			opt: &options{
				searchPath:   []string{"service-b"},
				searchGrep:   []string{"main"},
				indent:       defaultIndent,
				contextLines: 1,
			},
			expect: here.Doc(`
				testdata/service-b/
				testdata/service-b/main.go
				 2: 
				 3: func main() {
				 4: 	bar := 34
			`),
		},
		"service-b grep bar with C2": {
			opt: &options{
				searchPath:   []string{"service-b"},
				searchGrep:   []string{"main"},
				indent:       defaultIndent,
				contextLines: 2,
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
		},
		"service-c grep 56 with C2. Match 2 consecutive lines": {
			opt: &options{
				searchPath:   []string{"service-c"},
				searchGrep:   []string{"56"},
				indent:       defaultIndent,
				contextLines: 2,
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
		},
		"service-b grep onlyMatch": {
			opt: &options{
				searchPath:       []string{"service-b"},
				searchGrep:       []string{"func"},
				indent:           defaultIndent,
				onlyMatchContent: true,
			},
			expect: here.Doc(`
                testdata/service-b/main.go
                 3: func main() {
			`),
		},
		"service-b grep foo": {
			opt: &options{
				searchPath: []string{"service-a"},
				searchGrep: []string{"foo"},
				indent:     defaultIndent,
			},
			expect: here.Doc(`
                testdata/service-a/
                testdata/service-a/a.dat
                testdata/service-a/b
                testdata/service-a/main.go
                 4: 	foo := 12
			`),
		},
		"service-c grep foo": {
			opt: &options{
				searchPath: []string{"service-c"},
				searchGrep: []string{"foo"},
				indent:     defaultIndent,
			},
			expect: here.Doc(`
                testdata/service-c/
                testdata/service-c/main.go
                 7: 	foo()
                 10: func foo() {
			`),
		},
		"service-c grep foo noIndent": {
			opt: &options{
				searchPath: []string{"service-c"},
				searchGrep: []string{"foo"},
				indent:     defaultIndent,
				noIndent:   true,
			},
			expect: here.Doc(`
                testdata/service-c/
                testdata/service-c/main.go
                7: 	foo()
                10: func foo() {
			`),
		},
		"service-b grep custom indent string": {
			opt: &options{
				searchPath: []string{"service-b"},
				searchGrep: []string{"func"},
				indent:     "	",
			},
			expect: here.Doc(`
                testdata/service-b/
                testdata/service-b/main.go
                	3: func main() {
			`),
		},
		"service-d ignore .gitkeep": {
			opt: &options{
				searchPath: []string{"service-d"},
			},
			expect: here.Doc(`
                testdata/service-d/
			`),
		},
		"not pick .gitkeep even with --hidden option": {
			opt: &options{
				searchPath: []string{"service-d"},
				hidden:     true,
			},
			expect: here.Doc(`
                testdata/service-d/
			`),
		},
		"not pick dotfile by default": {
			opt: &options{
				searchPath: []string{"service-e"},
				hidden:     false, // default false
			},
			expect: here.Doc(`
                testdata/service-e/
			`),
		},
		"pick dotfile with --hidden option": {
			opt: &options{
				searchPath: []string{"service-e"},
				hidden:     true,
			},
			expect: here.Doc(`
                testdata/service-e/
                testdata/service-e/.config
			`),
		},
		"not pick up ignorez dir due to .gitignore": {
			opt: &options{
				searchPath: []string{"service-f"},
			},
			expect: here.Doc(`
                testdata/service-f/
			`),
		},
		"pick up ignorez dir with --skip-gitignore option": {
			opt: &options{
				searchPath:    []string{"service-f"},
				skipGitIgnore: true,
			},
			expect: here.Doc(`
                testdata/service-f/
                testdata/service-f/ignorez/
			`),
		},
		"ignore *min.js by default": {
			opt: &options{
				searchPath: []string{"service-g"},
			},
			expect: here.Doc(`
                testdata/service-g/
			`),
		},
		"ignore option": {
			opt: &options{
				searchPath: []string{"service-a"},
				ignore:     []string{"a.dat", "b"},
			},
			expect: here.Doc(`
                testdata/service-a/
                testdata/service-a/main.go
			`),
		},
		"ignore anyway even with --hidden option": {
			opt: &options{
				searchPath: []string{"service-e"},
				hidden:     true,
				ignore:     []string{".config"},
			},
			expect: here.Doc(`
                testdata/service-e/
			`),
		},
		"ignore anyway even with --search-all option": {
			opt: &options{
				searchPath: []string{"service-e"},
				searchAll:  true,
				ignore:     []string{".config"},
			},
			expect: here.Doc(`
                testdata/service-e/
			`),
		},
		"pick *min.js with --search-all option": {
			opt: &options{
				searchPath: []string{"service-g"},
				searchAll:  true,
			},
			expect: here.Doc(`
                testdata/service-g/
                testdata/service-g/service-g.min.js
			`),
		},
		"pick up ignorez dir with --search-all option": {
			opt: &options{
				searchPath: []string{"service-f"},
				searchAll:  true,
			},
			expect: here.Doc(`
                testdata/service-f/
                testdata/service-f/ignorez/
			`),
		},
		"service-h": {
			opt: &options{
				searchPath: []string{"service-h"},
				searchGrep: []string{"h"},
				indent:     defaultIndent,
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
		},
		"service-h with maxMatchCount": {
			opt: &options{
				searchPath:    []string{"service-h"},
				searchGrep:    []string{"h"},
				indent:        defaultIndent,
				maxMatchCount: 3,
			},
			expect: here.Doc(`
                testdata/service-h/
                testdata/service-h/main.go
                 1: package h
                 4: 	hi()
                 5: 	hello()
			`),
		},
		"service-h show count": {
			opt: &options{
				searchPath:     []string{"service-h"},
				searchGrep:     []string{"h"},
				showMatchCount: true,
			},
			expect: here.Doc(`
                testdata/service-h/
                testdata/service-h/main.go:5
			`),
		},
		"service-c with contextLines": {
			opt: &options{
				searchPath:     []string{"service-c"},
				searchGrep:     []string{"func"},
				groupSeparator: defaultGroupSeparator,
				indent:         defaultIndent,
				contextLines:   1,
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
		},
		"service-c with contextLines groupSeparator": {
			opt: &options{
				searchPath:     []string{"service-c"},
				searchGrep:     []string{"func"},
				groupSeparator: "====",
				indent:         defaultIndent,
				contextLines:   1,
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
		},
		"service-b ignore case to match": {
			opt: &options{
				searchPath: []string{"Service-B"},
				searchGrep: []string{"FunC"},
				indent:     defaultIndent,
				ignoreCase: true,
			},
			expect: here.Doc(`
                testdata/service-b/
                testdata/service-b/main.go
                 3: func main() {
			`),
		},
		"service-c with afterContextLines": {
			opt: &options{
				searchPath:        []string{"service-c"},
				searchGrep:        []string{"func"},
				groupSeparator:    defaultGroupSeparator,
				indent:            defaultIndent,
				afterContextLines: 1,
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
		},
		"service-c with beforeContextLines": {
			opt: &options{
				searchPath:         []string{"service-c"},
				searchGrep:         []string{"func"},
				groupSeparator:     defaultGroupSeparator,
				indent:             defaultIndent,
				beforeContextLines: 1,
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
		},
	} {
		t.Run(tname, func(t *testing.T) {
			var o bytes.Buffer
			cli := &runner{
				out:   &o,
				isTTY: true,
			}

			tt.opt.noPager = true
			tt.opt.noColor = true
			tt.opt.searchStart = "./testdata"

			cli.xfg(tt.opt)

			if os.Getenv("RUNNER_OS") == "Windows" {
				// BK: override path delimiter for Windows
				tt.expect = strings.ReplaceAll(tt.expect, "/", "\\")
			}

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
	    testdata/service-b/
	    testdata/service-b/main.go
	`)

	if os.Getenv("RUNNER_OS") == "Windows" {
		// BK: override path delimiter for Windows
		expect = strings.ReplaceAll(expect, "/", "\\")
	}

	a.Got(o.String()).Expect(expect).X().Same(t)
}
