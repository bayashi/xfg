package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	here "github.com/MakeNowJust/heredoc/v2"
	a "github.com/bayashi/actually"
	"github.com/bayashi/xfg/internal/xfgstats"
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
		out:   &outOutput,
		stats: xfgstats.New(1),
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
	t.Parallel()
	for tname, tt := range map[string]struct {
		opt            *options
		expect         string
		expectExitCode int
		skipWindows    bool
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
                testdata/service-b/main.go
                 3: func main() {
			`),
			expectExitCode: exitOK,
		},
		"grep `package b`": {
			opt: &options{
				SearchGrep: []string{"package b"},
				Indent:     defaultIndent,
			},
			expect: here.Doc(`
                testdata/service-b/main.go
                 1: package b
			`),
			expectExitCode: exitOK,
		},
		"service grep relax": {
			opt: &options{
				SearchPath: []string{"main"},
				SearchGrep: []string{"package"},
				Indent:     defaultIndent,
			},
			expect: here.Doc(`
                testdata/service-a/main.go
                 1: package a
                
                testdata/service-b/main.go
                 1: package b
                
                testdata/service-c/main.go
                 1: package c
                
                testdata/service-h/main.go
                 1: package h
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
				SearchPath: []string{"service-a"},
				SearchGrep: []string{"foo"},
				Indent:     defaultIndent,
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
		"not pick dot-dir by default": {
			opt: &options{
				SearchPath: []string{"service-o"},
				Hidden:     false, // default false
			},
			expect: here.Doc(`
                testdata/service-o/
			`),
			expectExitCode: exitOK,
		},
		"pick dot-dir with --hidden option": {
			opt: &options{
				SearchPath: []string{"service-o"},
				Hidden:     true,
			},
			expect: here.Doc(`
                testdata/service-o/
                testdata/service-o/.dotdir/
                testdata/service-o/.dotdir/foo.txt
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
		"ignore files and directories by default": {
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
		"pick *min.js etc with --search-all option": {
			opt: &options{
				SearchPath: []string{"service-g"},
				SearchAll:  true,
			},
			expect: here.Doc(`
                testdata/service-g/
                testdata/service-g/.svn/
                testdata/service-g/.svn/.gitkeep
                testdata/service-g/node_modules/
                testdata/service-g/node_modules/.gitkeep
                testdata/service-g/service-g.min.css
                testdata/service-g/service-g.min.js
                testdata/service-g/vendor/
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
		"service-b --files-with-matches": {
			opt: &options{
				SearchPath:       []string{"service-b"},
				FilesWithMatches: true,
			},
			expect: here.Doc(`
                testdata/service-b/main.go
			`),
			expectExitCode: exitOK,
		},
		"service grep bar --files-with-matches": {
			opt: &options{
				SearchPath:       []string{"service-b"},
				SearchGrep:       []string{"bar"},
				FilesWithMatches: true,
			},
			expect: here.Doc(`
                testdata/service-b/main.go
			`),
			expectExitCode: exitOK,
		},
		"service-b grep func --no-filename": {
			opt: &options{
				SearchPath: []string{"service-b"},
				SearchGrep: []string{"func"},
				NoFilename: true,
				Indent:     defaultIndent,
			},
			expect: " 3: func main() {\n",
			expectExitCode: exitOK,
		},
		"service grep package --no-filename": {
			opt: &options{
				SearchPath: []string{"main"},
				SearchGrep: []string{"package"},
				NoFilename: true,
				Indent:     defaultIndent,
			},
			expect: " 1: package a\n 1: package b\n 1: package c\n 1: package h\n",
			expectExitCode: exitOK,
		},
		"service-b --no-filename (ignored when path search only)": {
			opt: &options{
				SearchPath: []string{"service-b"},
				NoFilename: true,
			},
			expect: here.Doc(`
                testdata/service-b/
                testdata/service-b/main.go
			`),
			expectExitCode: exitOK,
		},
		"service-b --search-only-name": {
			opt: &options{
				SearchPath:     []string{"a.dat"},
				SearchOnlyName: true,
			},
			expect: here.Doc(`
                testdata/service-a/a.dat
			`),
			expectExitCode: exitOK,
		},
		"service-b --search-only-name service-b": {
			opt: &options{
				SearchPath:     []string{"service-b"},
				SearchOnlyName: true,
			},
			expect: here.Doc(`
                testdata/service-b/
			`),
			expectExitCode: exitOK,
		},
		"--search-only-name main.go": {
			opt: &options{
				SearchPath:     []string{"main.go"},
				SearchOnlyName: true,
			},
			expect: here.Doc(`
                testdata/service-a/main.go
                testdata/service-b/main.go
                testdata/service-c/main.go
                testdata/service-h/main.go
			`),
			expectExitCode: exitOK,
		},
		"ignore files and directories with onlyMatchContent by default": {
			opt: &options{
				SearchPath: []string{"service-j"},
				SearchGrep: []string{"foo"},
			},
			expect:         "", // skiped all
			expectExitCode: exitOK,
		},
		"search path service-(b|c)": {
			opt: &options{
				SearchPathRe: []string{"service-(b|c)$"},
			},
			expect: here.Doc(`
                testdata/service-b/
                testdata/service-c/
			`),
			expectExitCode: exitOK,
		},
		"search contents by regexp": {
			opt: &options{
				SearchGrepRe: []string{"ba(r|z) := \\d+$"},
			},
			expect: here.Doc(`
                testdata/service-b/main.go
                4: 	bar := 34
                
                testdata/service-c/main.go
                4: 	baz := 56
			`),
			expectExitCode: exitOK,
		},
		"search only perl files": {
			opt: &options{
				Lang: []string{"perl"},
			},
			expect: here.Doc(`
                testdata/service-k/bar.pl
                testdata/service-m/foo.pm
			`),
			expectExitCode: exitOK,
		},
		"search path by keyword, and filter language": {
			opt: &options{
				SearchPath: []string{"bar"},
				Lang:       []string{"perl"},
			},
			expect: here.Doc(`
                testdata/service-k/bar.pl
			`),
			expectExitCode: exitOK,
		},
		"search path by keyword, and filter language, and grep contents": {
			opt: &options{
				SearchPath: []string{"bar"},
				Lang:       []string{"perl"},
				SearchGrep: []string{"exit"},
			},
			expect: here.Doc(`
                testdata/service-k/bar.pl
                3: exit 0;
			`),
			expectExitCode: exitOK,
		},
		"search path by extension": {
			opt: &options{
				Ext: []string{"pl"},
			},
			expect: here.Doc(`
                testdata/service-k/bar.pl
			`),
			expectExitCode: exitOK,
		},
		"search path by extension with dot before extension name": {
			opt: &options{
				Ext: []string{".pl"},
			},
			expect: here.Doc(`
                testdata/service-k/bar.pl
			`),
			expectExitCode: exitOK,
		},
		"not match any words as word boundary regexp by default": {
			opt: &options{
				SearchGrepRe: []string{"bound"},
			},
			expect:         "",
			expectExitCode: exitOK,
		},
		"match words as word boundary regexp by default": {
			opt: &options{
				SearchGrepRe: []string{"boundary"},
			},
			expect: here.Doc(`
                testdata/service-n/foo.txt
                1: word boundary test line
			`),
			expectExitCode: exitOK,
		},
		"just match line as not word boundary regexp": {
			opt: &options{
				SearchGrepRe:    []string{"bound"},
				NotWordBoundary: true,
			},
			expect: here.Doc(`
                testdata/service-n/foo.txt
                1: word boundary test line
			`),
			expectExitCode: exitOK,
		},
		"--type d service-d": {
			opt: &options{
				SearchPath: []string{"service-d"},
				Type:       "d",
			},
			expect: here.Doc(`
                testdata/service-d/
			`),
			expectExitCode: exitOK,
		},
		"--type l": {
			opt: &options{
				SearchPath: []string{"service-p"},
				Type:       "l",
			},
			expect: here.Doc(`
                testdata/service-p/testlink
			`),
			expectExitCode: exitOK,
		},
		"--type x": {
			opt: &options{
				SearchPath: []string{"service-p"},
				Type:       "x",
			},
			expect: here.Doc(`
                testdata/service-p/a.sh
			`),
			expectExitCode: exitOK,
			skipWindows:    true, // Windows doesn't have executable permission
		},
		"--type e": {
			opt: &options{
				SearchPath: []string{"service-k"},
				Type:       "e",
			},
			expect: here.Doc(`
                testdata/service-k/foo.txt
			`),
			expectExitCode: exitOK,
		},
		"pick *min.js etc with --hidden and --no-default-skip": {
			opt: &options{
				SearchPath:    []string{"service-q"},
				Hidden:        true,
				NoDefaultSkip: true,
			},
			expect: here.Doc(`
                testdata/service-q/
                testdata/service-q/.svn/
                testdata/service-q/.svn/.gitkeep
                testdata/service-q/node_modules/
                testdata/service-q/node_modules/.gitkeep
                testdata/service-q/service-q.min.css
                testdata/service-q/service-q.min.js
			`),
			expectExitCode: exitOK,
		},
		"unicode emoji": {
			opt: &options{
				SearchPath: []string{"service-r"},
				SearchGrep: []string{"😂"},
			},
			expect: here.Doc(`
                testdata/service-r/emoji.txt
                3: 😂
			`),
			expectExitCode: exitOK,
		},
		"multiple start dirs": {
			opt: &options{
				SearchStart: []string{"./testdata/service-a", "./testdata/service-b"},
				SearchPath:  []string{"main"},
			},
			expect: here.Doc(`
                testdata/service-a/main.go
                testdata/service-b/main.go
			`),
			expectExitCode: exitOK,
		},
		"multiple start dirs grep": {
			opt: &options{
				SearchStart: []string{"./testdata/service-a", "./testdata/service-b"},
				SearchPath:  []string{"main"},
				SearchGrep:  []string{"package b"},
			},
			expect: here.Doc(`
                testdata/service-b/main.go
                1: package b
			`),
			expectExitCode: exitOK,
		},
		"pick *min.js etc with --search-default-skip-stuff": {
			opt: &options{
				SearchPath:             []string{"service-q"},
				SearchDefaultSkipStuff: true,
			},
			expect: here.Doc(`
                testdata/service-q/
                testdata/service-q/.svn/
                testdata/service-q/.svn/.gitkeep
                testdata/service-q/node_modules/
                testdata/service-q/node_modules/.gitkeep
                testdata/service-q/service-q.min.css
                testdata/service-q/service-q.min.js
			`),
			expectExitCode: exitOK,
		},
		"Not pick up any due to maxDepth": {
			opt: &options{
				SearchPath: []string{"service-s"},
				SearchGrep: []string{"bar"},
				MaxDepth:   2,
			},
			expect:         "",
			expectExitCode: exitOK,
		},
		"Pick up d3, however, not pick up d4.txt due to maxDepth": {
			opt: &options{
				SearchPath: []string{"service-s"},
				SearchGrep: []string{"bar"},
				MaxDepth:   3,
			},
			expect: here.Doc(`
                testdata/service-s/d3/d3.txt
                1: bar
			`),
			expectExitCode: exitOK,
		},
		"Pick up until d4 by enough maxDepth": {
			opt: &options{
				SearchPath: []string{"service-s"},
				SearchGrep: []string{"bar"},
				MaxDepth:   4,
			},
			expect: here.Doc(`
                testdata/service-s/d3/d3.txt
                1: bar
                
                testdata/service-s/d3/d4/d4.txt
                1: bar
			`),
			expectExitCode: exitOK,
		},
	} {
		if tt.skipWindows && isWindowsTestRunner() {
			return
		}

		tt := tt
		t.Run(tname, func(t *testing.T) {
			t.Parallel()
			var o bytes.Buffer
			cli := &runner{
				out:   &o,
				isTTY: true,
				stats: xfgstats.New(1),
			}

			tt.opt.NoPager = true
			tt.opt.NoColor = true
			if tt.opt.SearchStart == nil {
				tt.opt.SearchStart = []string{"./testdata"}
			}
			if tt.opt.MaxDepth == 0 {
				tt.opt.MaxDepth = defaultMaxDepth
			}
			tt.opt.KeepResultOrder = true

			code, err := cli.xfg(tt.opt)

			a.Got(err).Debug("options", tt.opt).NoError(t)
			a.Got(code).Expect(tt.expectExitCode).Same(t)

			tt.expect = windowsBK(tt.expect)
			a.Got(o.String()).Expect(tt.expect).X().Debug("options", tt.opt).Same(t)
		})
	}
}

// no color, no pager
func TestNonTTY(t *testing.T) {
	for tname, tt := range map[string]struct {
		args           []string
		expect         string
		expectExitCode int
	}{
		"service-b func": {
			args: []string{"service-b", "func"},
			expect: here.Doc(`
			    testdata/service-b/main.go:3:func main() {
			`),
			expectExitCode: exitOK,
		},
		"service-b func --files-with-matches": {
			args: []string{"service-b", "func", "--files-with-matches"},
			expect: here.Doc(`
			    testdata/service-b/main.go
			`),
			expectExitCode: exitOK,
		},
		"service-b func --files-with-matches --null": {
			args:           []string{"service-b", "func", "--files-with-matches", "--null"},
			expect:         "testdata/service-b/main.go\x00",
			expectExitCode: exitOK,
		},
	} {
		t.Run(tname, func(t *testing.T) {
			resetFlag()
			stubExit()
			os.Args = append([]string{fakeCmd, "-s", "./testdata"}, tt.args...)
			var o bytes.Buffer
			cli := &runner{
				out:   &o,
				isTTY: false,
				stats: xfgstats.New(1),
			}

			exitCode, msg := cli.run()
			a.Got(msg).Expect("").Same(t)
			a.Got(exitCode).Expect(exitOK).Same(t)

			actualExpect := windowsBK(tt.expect)

			a.Got(o.String()).Expect(actualExpect).X().Same(t)
		})
	}
}
