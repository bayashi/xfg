package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	here "github.com/MakeNowJust/heredoc/v2"
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
		"only path regexp arg": {
			args: []string{"-P", "fo."},
			prepareExpect: func(o *options) {
				o.SearchPathRe = []string{"fo."}
			},
		},
		"only grep regexp arg": {
			args: []string{"-G", "fo."},
			prepareExpect: func(o *options) {
				o.SearchGrepRe = []string{"fo."}
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

func TestShowLangList(t *testing.T) {
	expect := here.Doc(`
		ada: .ada, .adb, .ads
		asciidoc: .asc, .asciidoc, .adoc, .ad
		asm: .asm, .s, .S
		asp: .asmx, .aspx, .aspx.cs, .asax, .asp, .asa, .ashx, .ascx, .ascx.cs, .aspx.vb, .ascx.vb
		aspx: .ashx, .ascx, .asp, .asa, .aspx, .asax, .asmx
		batch: .bat, .cmd
		bazel: .bazel, .bzl, .BUILD, .bazelrc, BUILD, MODULE.bazel, WORKSPACE, WORKSPACE.bazel
		bitbake: .bb, .bbappend, .bbclass, .inc, .conf
		cc: .xs, .c, .h
		cfmx: .cfc, .cfm, .cfml
		clojure: .edn, .cljc, .cljx, .clj, .cljs
		coffee: .coffee, .cjsx
		coq: .coq, .g, .v
		cpp: .tpp, .m, .hpp, .H, .hxx, .C, .cxx, .hh, .h, .cpp, .cc
		css: .css, sass, .scss
		cython: .pxi, .pyx, .pxd
		delphi: .bdsproj, .pas, .int, .dpr, .dproj, .dfm, .nfm, .dof, .dpk, .groupproj, .bdsgroup
		ebuild: .ebuild, .eclass
		elixir: .ex, .eex, .exs
		erlang: .erl, .hrl
		fortran: .F90, .f95, .ftn, .fpp, .f77, .f90, .f, .F, .f03, .for, .FPP
		fsharp: .fsx, .fs, .fsi
		gettext: .mo, .po, .pot
		glsl: .vert, .tesc, .tese, .geom, .frag, .comp
		groovy: .gradle, .gpp, .grunit, .groovy, .gtmpl
		haskell: .hs, .hsig, .lhs
		html: .shtml, .xhtml, .htm, .html
		idris: .lidr, .idr, .ipkg
		java: .java, .properties
		js: .es6, .js, .jsx, .vue
		jsp: .jsp, .jspx, .tagf, .jhtm, .jhtml, .jspf, .tag
		make: .Makefiles, .mk, .mak
		markdown: .mdwn, .mkdn, .markdown, .mdown, .mkd, .md
		mason: .mas, .mhtml, .mpl, .mtxt
		md: .mkd, .md, .markdown, .mdown, .mdwn, .mkdn
		ocaml: .ml, .mli, .mll, .mly
		parrot: .pir, .pasm, .pmc, .ops, .pod, .pg, .tg
		perl: .pl, .pm, .t, .pod, .PL
		php: .php3, .php4, .php5, .phtml, .php, .phpt
		plone: .metadata, .cpy, .zcml, .py, .xml, .pt, .cpt
		r: .r, .R, .Rtex, .Rrst, .Rmd, .Rnw
		racket: .scm, .rkt, .ss
		ruby: .rjs, .rxml, .erb, .rake, .spec, .rb, .rhtml
		shell: .csh, .tcsh, .ksh, .zsh, .sh, .bash, .fish
		sml: .sml, .fun, .mlb, .sig
		tcl: .tcl, .itcl, .itk
		tex: .sty, .tex, .cls
		ts: .ts, .tsx
		tt: .tt, .tt2, .ttml
		vala: .vala, .vapi
		vb: .bas, .cls, .vb, .resx, .frm, .ctl
		velocity: .vsl, .vm, .vtl
		verilog: .sv, .svh, .v, .vh
		vhdl: .vhd, .vhdl
		xml: .xsd, .ent, .xsl, .xslt, .tld, .plist, .wsdl, .xml, .dtd
		yaml: .yaml, .yml
		yml: .yml, .yaml
		zeek: .bif, .zeek, .bro
	`)
	var errOutput bytes.Buffer
	cli := &runner{
		err: &errOutput,
	}
	resetFlag()
	stubExit()
	os.Args = []string{fakeCmd, "--lang-list"}
	o := cli.parseArgs(defaultOptions())
	e := expectedDefaultOptions()
	a.Got(o).Expect(&e).Same(t)
	a.Got(stubCalled).True(t)
	a.Got(stubCode).Expect(exitOK).Same(t)
	a.Got(errOutput.String()).Expect(expect).Same(t)
}
