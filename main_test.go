package main

import (
	"bytes"
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
				path: "service-b",
			},
			expect: here.Doc(`
                testdata/service-b
                testdata/service-b/main.go
			`),
		},
		"service-b grep": {
			opt: &options{
				path: "service-b",
				grep: "func",
			},
			expect: here.Doc(`
                testdata/service-b
                testdata/service-b/main.go
                  3: func main() {}
			`),
		},
		"service grep relax": {
			opt: &options{
				path:  "main",
				grep:  "package b",
				relax: true,
			},
			expect: here.Doc(`
                testdata/service-a/main.go
                testdata/service-b/main.go
                  1: package b

                testdata/service-c/main.go
			`),
		},
	} {
		t.Run(tname, func(t *testing.T) {
			var o bytes.Buffer
			cli := &runner{
				out: &o,
			}

			tt.opt.start = "./testdata"

			cli.xfg(tt.opt)

			a.Got(o.String()).Expect(tt.expect).X().Same(t)
		})
	}
}
