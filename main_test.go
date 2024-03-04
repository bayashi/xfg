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
			},
			expect: here.Doc(`
                testdata/service-b
                testdata/service-b/main.go
			`),
		},
		"service-b grep": {
			opt: &options{
				searchPath: "service-b",
				searchGrep: "func",
			},
			expect: here.Doc(`
                testdata/service-b
                testdata/service-b/main.go
                  3: func main() {
			`),
		},
		"service grep relax": {
			opt: &options{
				searchPath: "main",
				searchGrep: "package b",
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
				contextLines: 1,
			},
			expect: here.Doc(`
				testdata/service-b
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
				contextLines: 2,
			},
			expect: here.Doc(`
				testdata/service-b
				testdata/service-b/main.go
				  1: package b
				  2: 
				  3: func main() {
				  4: 	bar := 34
				  5: }
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
