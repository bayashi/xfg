package main

import (
	"bytes"
	"testing"

	a "github.com/bayashi/actually"
	"github.com/bayashi/xfg/internal/xfgstats"
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
				SearchPath: []string{"service-b"},
			},
			expect: "\x1b[93mtestdata/\x1b[96mservice-b\x1b[0m\x1b[93m/\x1b[0m\n" +
				"\x1b[93mtestdata/\x1b[96mservice-b\x1b[0m\x1b[93m/main.go\x1b[0m\n",
		},
		"service-b color green": {
			opt: &options{
				SearchPath: []string{"service-b"},
				ColorPath:  "green",
			},
			expect: "\x1b[93mtestdata/\x1b[92mservice-b\x1b[0m\x1b[93m/\x1b[0m\n" +
				"\x1b[93mtestdata/\x1b[92mservice-b\x1b[0m\x1b[93m/main.go\x1b[0m\n",
		},
		"service-b grep": {
			opt: &options{
				SearchPath: []string{"service-b"},
				SearchGrep: []string{"func"},
				Indent:     defaultIndent,
			},
			expect: "\x1b[93mtestdata/\x1b[96mservice-b\x1b[0m\x1b[93m/\x1b[0m\n" +
				"\x1b[93mtestdata/\x1b[96mservice-b\x1b[0m\x1b[93m/main.go\x1b[0m\n" +
				" \x1b[91m3\x1b[0m: \x1b[91mfunc\x1b[0m main() {\n",
		},
		"service-b grep green": {
			opt: &options{
				SearchPath:   []string{"service-b"},
				SearchGrep:   []string{"func"},
				Indent:       defaultIndent,
				ColorContent: "green",
			},
			expect: "\x1b[93mtestdata/\x1b[96mservice-b\x1b[0m\x1b[93m/\x1b[0m\n" +
				"\x1b[93mtestdata/\x1b[96mservice-b\x1b[0m\x1b[93m/main.go\x1b[0m\n" +
				" \x1b[92m3\x1b[0m: \x1b[92mfunc\x1b[0m main() {\n",
		},
		"serv ice-b": {
			opt: &options{
				SearchPath: []string{"ser", "ice-b"},
			},
			expect: "\x1b[93mtestdata/\x1b[96mser\x1b[0m\x1b[93mv\x1b[96mice-b\x1b[0m\x1b[93m/\x1b[0m\n" +
				"\x1b[93mtestdata/\x1b[96mser\x1b[0m\x1b[93mv\x1b[96mice-b\x1b[0m\x1b[93m/main.go\x1b[0m\n",
		},
		"service-b grep multiple keywords": {
			opt: &options{
				SearchPath: []string{"service-b"},
				SearchGrep: []string{"func", "main"},
				Indent:     defaultIndent,
			},
			expect: "\x1b[93mtestdata/\x1b[96mservice-b\x1b[0m\x1b[93m/\x1b[0m\n" +
				"\x1b[93mtestdata/\x1b[96mservice-b\x1b[0m\x1b[93m/main.go\x1b[0m\n" +
				" \x1b[91m3\x1b[0m: \x1b[91mfunc\x1b[0m \x1b[91mmain\x1b[0m() {\n",
		},
		"service-b path base color red": {
			opt: &options{
				SearchPath:    []string{"service-b"},
				ColorPathBase: "red",
			},
			expect: "\x1b[91mtestdata/\x1b[96mservice-b\x1b[0m\x1b[91m/\x1b[0m\n" +
				"\x1b[91mtestdata/\x1b[96mservice-b\x1b[0m\x1b[91m/main.go\x1b[0m\n",
		},
	} {
		t.Run(tname, func(t *testing.T) {
			var o bytes.Buffer
			cli := &runner{
				out:   &o,
				isTTY: true,
				stats: xfgstats.New(1),
			}

			tt.opt.NoPager = true
			tt.opt.SearchStart = "./testdata"

			cli.xfg(tt.opt)

			tt.expect = windowsBK(tt.expect)

			a.Got(o.String()).Expect(tt.expect).X().Debug("options", tt.opt).Same(t)
		})
	}
}

func TestNoColorByENV(t *testing.T) {
	for tname, tt := range map[string]struct {
		opt    *options
		expect string
	}{
		"service-b": {
			opt: &options{
				SearchPath: []string{"service-b"},
			},
			expect: "testdata/service-b/\n" +
				"testdata/service-b/main.go\n",
		},
	} {
		t.Run(tname, func(t *testing.T) {
			var o bytes.Buffer
			cli := &runner{
				out:   &o,
				isTTY: true,
				stats: xfgstats.New(1),
			}

			tt.opt.NoPager = true
			tt.opt.SearchStart = "./testdata"

			t.Setenv("NO_COLOR", "1")

			cli.xfg(tt.opt)

			tt.expect = windowsBK(tt.expect)

			a.Got(o.String()).Expect(tt.expect).X().Debug("options", tt.opt).Same(t)
		})
	}
}
