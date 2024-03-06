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
