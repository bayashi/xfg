package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	a "github.com/bayashi/actually"
	c "github.com/fatih/color"
)

func init() {
	c.NoColor = false
}

func TestHighlight(t *testing.T) {
	for tname, tt := range map[string]struct {
		opt    *options
		expect string
	}{
		"service-b": {
			opt: &options{
				searchPath: "service-b",
			},
			expect: "testdata/\x1b[96mservice-b\x1b[0m/\n" +
				"testdata/\x1b[96mservice-b\x1b[0m/main.go\n",
		},
		"service-b color green": {
			opt: &options{
				searchPath: "service-b",
				colorPath:  "green",
			},
			expect: "testdata/\x1b[92mservice-b\x1b[0m/\n" +
				"testdata/\x1b[92mservice-b\x1b[0m/main.go\n",
		},
		"service-b grep": {
			opt: &options{
				searchPath: "service-b",
				searchGrep: "func",
				indent:     defaultIndent,
			},
			expect: "testdata/\x1b[96mservice-b\x1b[0m/\n" +
				"testdata/\x1b[96mservice-b\x1b[0m/main.go\n" +
				" \x1b[91m3\x1b[0m: \x1b[91mfunc\x1b[0m main() {\n",
		},
		"service-b grep green": {
			opt: &options{
				searchPath:   "service-b",
				searchGrep:   "func",
				indent:       defaultIndent,
				colorContent: "green",
			},
			expect: "testdata/\x1b[96mservice-b\x1b[0m/\n" +
				"testdata/\x1b[96mservice-b\x1b[0m/main.go\n" +
				" \x1b[92m3\x1b[0m: \x1b[92mfunc\x1b[0m main() {\n",
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

			a.Got(o.String()).Expect(tt.expect).X().Debug("options", tt.opt).Same(t)
		})
	}
}
