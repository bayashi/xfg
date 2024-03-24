package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	here "github.com/MakeNowJust/heredoc/v2"
	a "github.com/bayashi/actually"
)

func TestRunner_OK(t *testing.T) {
	for tname, tt := range map[string]struct {
		opt    *options
		expect string
	}{
		"service-b": {
			opt: &options{
				searchPath: "service-b",
				indent:     "  ",
			},
			expect: here.Doc(`
                testdata/service-b/
                testdata/service-b/main.go
			`),
		},
		"service-b grep": {
			opt: &options{
				searchPath: "service-b",
				searchGrep: "func",
				indent:     "  ",
			},
			expect: here.Doc(`
                testdata/service-b/
                testdata/service-b/main.go
                  3: func main() {
			`),
		},
		"service grep relax": {
			opt: &options{
				searchPath: "main",
				searchGrep: "package b",
				indent:     "  ",
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
				searchPath:   "service-b",
				searchGrep:   "main",
				indent:       "  ",
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
				searchPath:   "service-b",
				searchGrep:   "main",
				indent:       "  ",
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
				searchPath:   "service-c",
				searchGrep:   "56",
				indent:       "  ",
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
				searchPath: "service-b",
				searchGrep: "func",
				indent:     "  ",
				onlyMatch:  true,
			},
			expect: here.Doc(`
                testdata/service-b/main.go
                  3: func main() {
			`),
		},
		"service-b grep foo": {
			opt: &options{
				searchPath: "service-a",
				searchGrep: "foo",
				indent:     "  ",
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
				searchPath:     "service-c",
				searchGrep:     "foo",
				groupSeparator: "--",
				indent:         "  ",
			},
			expect: here.Doc(`
                testdata/service-c/
                testdata/service-c/main.go
                  7: 	foo()
                  --
                  10: func foo() {
			`),
		},
		"service-c grep foo noIndent": {
			opt: &options{
				searchPath:     "service-c",
				searchGrep:     "foo",
				groupSeparator: "--",
				indent:         "  ",
				noIndent:       true,
			},
			expect: here.Doc(`
                testdata/service-c/
                testdata/service-c/main.go
                7: 	foo()
                --
                10: func foo() {
			`),
		},
		"service-b grep custom indent string": {
			opt: &options{
				searchPath: "service-b",
				searchGrep: "func",
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
				searchPath: "service-d",
			},
			expect: here.Doc(`
                testdata/service-d/
			`),
		},
		"not pick .gitkeep even with --hidden option": {
			opt: &options{
				searchPath: "service-d",
				hidden:     true,
			},
			expect: here.Doc(`
                testdata/service-d/
			`),
		},
		"not pick dotfile by default": {
			opt: &options{
				searchPath: "service-e",
				hidden:     false, // default false
			},
			expect: here.Doc(`
                testdata/service-e/
			`),
		},
		"pick dotfile with --hidden option": {
			opt: &options{
				searchPath: "service-e",
				hidden:     true,
			},
			expect: here.Doc(`
                testdata/service-e/
                testdata/service-e/.config
			`),
		},
		"not pick up ignorez dir due to .gitignore": {
			opt: &options{
				searchPath: "service-f",
			},
			expect: here.Doc(`
                testdata/service-f/
			`),
		},
		"pick up ignorez dir with --skip-gitignore option": {
			opt: &options{
				searchPath:    "service-f",
				skipGitIgnore: true,
			},
			expect: here.Doc(`
                testdata/service-f/
                testdata/service-f/ignorez/
			`),
		},
		"ignore *min.js by default": {
			opt: &options{
				searchPath: "service-g",
			},
			expect: here.Doc(`
                testdata/service-g/
			`),
		},
		"ignore option": {
			opt: &options{
				searchPath: "service-a",
				ignore:     []string{"a.dat", "b"},
			},
			expect: here.Doc(`
                testdata/service-a/
                testdata/service-a/main.go
			`),
		},
		"ignore anyway even with --hidden option": {
			opt: &options{
				searchPath: "service-e",
				hidden:     true,
				ignore:     []string{".config"},
			},
			expect: here.Doc(`
                testdata/service-e/
			`),
		},
		"ignore anyway even with --search-all option": {
			opt: &options{
				searchPath: "service-e",
				searchAll:  true,
				ignore:     []string{".config"},
			},
			expect: here.Doc(`
                testdata/service-e/
			`),
		},
		"pick *min.js with --search-all option": {
			opt: &options{
				searchPath: "service-g",
				searchAll:  true,
			},
			expect: here.Doc(`
                testdata/service-g/
                testdata/service-g/service-g.min.js
			`),
		},
		"pick up ignorez dir with --search-all option": {
			opt: &options{
				searchPath: "service-f",
				searchAll:  true,
			},
			expect: here.Doc(`
                testdata/service-f/
                testdata/service-f/ignorez/
			`),
		},
		"service-h": {
			opt: &options{
				searchPath:     "service-h",
				searchGrep:     "h",
				groupSeparator: "--",
				indent:         "  ",
			},
			expect: here.Doc(`
                testdata/service-h/
                testdata/service-h/main.go
                  1: package h
                  --
                  4: 	hi()
                  5: 	hello()
                  --
                  8: func hi() {
                  --
                  11: func hello() {
			`),
		},
		"service-h with maxMatchCount": {
			opt: &options{
				searchPath:     "service-h",
				searchGrep:     "h",
				groupSeparator: "--",
				indent:         "  ",
				maxMatchCount:  3,
			},
			expect: here.Doc(`
                testdata/service-h/
                testdata/service-h/main.go
                  1: package h
                  --
                  4: 	hi()
                  5: 	hello()
			`),
		},
		"service-h show count": {
			opt: &options{
				searchPath:     "service-h",
				searchGrep:     "h",
				showMatchCount: true,
			},
			expect: here.Doc(`
                testdata/service-h/
                testdata/service-h/main.go:5
			`),
		},
	} {
		t.Run(tname, func(t *testing.T) {
			var o bytes.Buffer
			cli := &runner{
				out: &o,
			}

			tt.opt.searchStart = "./testdata"

			cli.xfg(tt.opt)

			if os.Getenv("RUNNER_OS") == "Windows" {
				// BK: override path delimiter for Windows
				tt.expect = strings.ReplaceAll(tt.expect, "/", "\\")
			}

			a.Got(o.String()).Expect(tt.expect).X().Same(t)
		})
	}
}
