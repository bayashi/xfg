package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/bayashi/xfg/internal/xfglangxt"
	flag "github.com/spf13/pflag"
)

const (
	cmdName string = "xfg"

	XFG_RC_FILE string = ".xfgrc"

	defaultGroupSeparator string = "--"
	defaultIndent         string = " "

	supportTypes = "directory (d), symlink (l), executable (x), empty (e), socket (s), pipe (p), block-device (b), char-device (c)"
)

var (
	version     = ""
	installFrom = "Source"
)

// runtime options
type optionsExtra struct {
	actualAfterContextLines  uint32
	actualBeforeContextLines uint32
	withAfterContextLines    bool
	withBeforeContextLines   bool

	runWithNoArg bool

	onlyMatchContent bool
}

type options struct {
	SearchPath  []string `toml:"path"`
	SearchGrep  []string `toml:"grep"`
	SearchStart string   `toml:"start"`

	SearchPathRe []string `toml:"path-regexp"`
	SearchGrepRe []string `toml:"grep-regexp"`

	GroupSeparator string `toml:"gourp-separator"`
	Indent         string `toml:"indent"`
	ColorPathBase  string `toml:"color-path-base"`
	ColorPath      string `toml:"color-path"`
	ColorContent   string `toml:"color-conetnt"`
	XfgIgnoreFile  string `toml:"xfgignore-file"`

	Ignore []string `toml:"ignore"`

	Type string   `toml:"Type"`
	Lang []string `toml:"lang"`
	Ext  []string `toml:"ext"`

	IgnoreCase            bool `toml:"ignore-case"`
	NoColor               bool `toml:"no-color"`
	Abs                   bool `toml:"abs"`
	ShowMatchCount        bool `toml:"count"`
	NoGroupSeparator      bool `toml:"no-group-separator"`
	NoIndent              bool `toml:"no-indent"`
	Hidden                bool `toml:"hidden"`
	SkipGitIgnore         bool `toml:"skip-gitignore"`
	SkipXfgIgnore         bool `toml:"skip-xfgignore"`
	NoDefaultSkip         bool `toml:"no-default-skip"`
	SearchAll             bool `toml:"search-all"`
	Unrestricted          bool `toml:"unrestricted"`
	NoPager               bool `toml:"no-pager"`
	Quiet                 bool `toml:"quiet"`
	FilesWithMatches      bool `toml:"files-with-matches"`
	Null                  bool `toml:"null"`
	Stats                 bool `toml:"stats"`
	SearchOnlyName        bool `toml:"search-only-name"`
	NotWordBoundary       bool `toml:"not-word-boundary"`
	IgnorePermissionError bool `toml:"ignore-permission-error"`

	flagLangList bool

	ContextLines uint32 `toml:"context"`

	AfterContextLines  uint32 `toml:"after-context"`
	BeforeContextLines uint32 `toml:"before-context"`

	MaxMatchCount uint32 `toml:"max-count"`
	MaxColumns    uint32 `toml:"max-columns"`

	extra optionsExtra
}

func (o *options) falgs(d *options) {
	flag.StringArrayVarP(&o.SearchPath, "path", "p", d.SearchPath, "A string to find paths")
	flag.StringArrayVarP(&o.SearchGrep, "grep", "g", d.SearchGrep, "A string to search for contents")
	flag.StringVarP(&o.SearchStart, "start", "s", d.SearchStart, "A location to start searching")

	flag.BoolVarP(&o.IgnoreCase, "ignore-case", "i", d.IgnoreCase, "Ignore case distinctions to search. Also affects keywords of ignore option")

	flag.StringArrayVarP(&o.SearchPathRe, "path-regexp", "P", d.SearchPathRe, "A string to find paths by regular expressions (RE2)")
	flag.StringArrayVarP(&o.SearchGrepRe, "grep-regexp", "G", d.SearchGrepRe, "A string to grep contents by regular expressions (RE2)")
	flag.BoolVarP(&o.NotWordBoundary, "not-word-boundary", "", d.NotWordBoundary, "Not care about word boundary to match by regexp")

	flag.Uint32VarP(&o.ContextLines, "context", "C", d.ContextLines, "Show several lines before and after the matched one")
	flag.Uint32VarP(&o.AfterContextLines, "after-context", "A", d.AfterContextLines, "Show several lines after the matched one. Override context option")
	flag.Uint32VarP(&o.BeforeContextLines, "before-context", "B", d.BeforeContextLines, "Show several lines before the matched one. Override context option")

	flag.BoolVarP(&o.Hidden, "hidden", ".", d.Hidden, "Enable to search hidden files")
	flag.BoolVarP(&o.SkipGitIgnore, "skip-gitignore", "", d.SkipGitIgnore, "Search files and directories even if a path matches a line of .gitignore")
	flag.BoolVarP(&o.SkipXfgIgnore, "skip-xfgignore", "", d.SkipXfgIgnore, "Search files and directories even if a path matches a line of .xfgignore")
	flag.BoolVarP(&o.NoDefaultSkip, "no-default-skip", "", d.NoDefaultSkip, "Not skip .git, .gitkeep, .gitkeep, .svn, node_modules, vendor, *.min.js and *.mmin.css")
	flag.BoolVarP(&o.SearchAll, "search-all", "a", d.SearchAll, "Search all files and directories except specific ignoring files and directories")
	flag.BoolVarP(&o.Unrestricted, "unrestricted", "u", d.Unrestricted, "The alias of --search-all")
	flag.StringArrayVarP(&o.Ignore, "ignore", "", d.Ignore, "Ignore path to pick up even with '--search-all'")
	flag.BoolVarP(&o.SearchOnlyName, "search-only-name", "f", d.SearchOnlyName, "Search to only name instead whole path string")

	flag.StringVarP(&o.Type, "type", "t", d.Type, "Filter by file type: "+supportTypes)
	flag.StringArrayVarP(&o.Ext, "ext", "", d.Ext, "Only search files matching file extension")
	flag.StringArrayVarP(&o.Lang, "lang", "", d.Lang, "Only search files matching language. --lang-list prints all support languages")
	flag.BoolVarP(&o.flagLangList, "lang-list", "", false, "Show all supported file extensions for each language")

	flag.BoolVarP(&o.Abs, "abs", "", d.Abs, "Show absolute paths")
	flag.BoolVarP(&o.ShowMatchCount, "count", "c", d.ShowMatchCount, "Show a count of matching lines instead of contents")
	flag.Uint32VarP(&o.MaxMatchCount, "max-count", "m", d.MaxMatchCount, "Stop reading a file after NUM matching lines")
	flag.Uint32VarP(&o.MaxColumns, "max-columns", "", d.MaxColumns, "Do not print lines longer than this limit")
	flag.BoolVarP(&o.FilesWithMatches, "files-with-matches", "l", d.FilesWithMatches, "Only print the names of matching files")
	flag.BoolVarP(&o.Null, "null", "0", d.Null, "Separate the filenames with \\0, rather than \\n")

	flag.BoolVarP(&o.NoColor, "no-color", "", d.NoColor, "Disable colors for an output")
	flag.StringVarP(&o.ColorPathBase, "color-path-base", "", d.ColorPathBase, "Color name for a path")
	flag.StringVarP(&o.ColorPath, "color-path", "", d.ColorPath, "Color name to highlight keywords in a path")
	flag.StringVarP(&o.ColorContent, "color-content", "", d.ColorContent, "Color name to highlight keywords in a content line")

	flag.StringVarP(&o.GroupSeparator, "group-separator", "", d.GroupSeparator, "Print this string instead of '--' between groups of lines")
	flag.BoolVarP(&o.NoGroupSeparator, "no-group-separator", "", d.NoGroupSeparator, "Do not print a separator between groups of lines")

	flag.StringVarP(&o.Indent, "indent", "", d.Indent, "Indent string for the top of each line")
	flag.BoolVarP(&o.NoIndent, "no-indent", "", d.NoIndent, "Do not print an indent string")

	flag.BoolVarP(&o.IgnorePermissionError, "ignore-permission-error", "", d.IgnorePermissionError, "Do not print warnings of file permission error")

	flag.StringVarP(&o.XfgIgnoreFile, "xfgignore-file", "", d.XfgIgnoreFile, ".xfgignore file path if you have it except XDG base directory or HOME directory")

	flag.BoolVarP(&o.NoPager, "no-pager", "", d.NoPager, "Do not invoke with the Pager")
	flag.BoolVarP(&o.Quiet, "quiet", "q", d.Quiet, "Do not write anything to standard output. Exit immediately with zero status if any match is found")
	flag.BoolVarP(&o.Stats, "stats", "", d.Stats, "Print runtime stats after searching result")
}

func (cli *runner) parseArgs(d *options) *options {
	o := &options{}
	o.extra.runWithNoArg = len(os.Args) == 1
	o.falgs(d)

	flag.CommandLine.SetOutput(cli.err)
	flag.CommandLine.SortFlags = false
	var flagHelp, flagVersion bool
	flag.BoolVarP(&flagHelp, "help", "h", false, "Show help (This message) and exit")
	flag.BoolVarP(&flagVersion, "version", "v", false, "Show version and build command info and exit")
	flag.Parse()

	if o.Type != "" && !o.validateType() {
		cli.putErr(fmt.Sprintf("wrong type `%s`. Supported: %s", o.Type, supportTypes))
		funcExit(exitErr)
	}

	if flagHelp {
		cli.putHelp(fmt.Sprintf("Version %s", getVersion()))
	} else if flagVersion {
		cli.putErr(versionDetails())
		funcExit(exitOK)
	} else if o.flagLangList {
		cli.putErr(showLangList())
		funcExit(exitOK)
	} else {
		o.targetPathFromArgs()
	}

	return o
}

func (o *options) validateType() bool {
	if len(o.Type) == 1 && strings.Contains("fdlxespbc", o.Type) {
		return true // fine!
	}
	if o.Type == "directory" || o.Type == "symlink" || o.Type == "executable" || o.Type == "empty" ||
		o.Type == "socket" || o.Type == "pipe" || o.Type == "block-device" || o.Type == "char-device" {
		return true // fine!
	}

	return false
}

func (o *options) targetPathFromArgs() {
	if len(flag.Args()) > 0 && flag.Args()[0] != "" {
		o.SearchPath = append(o.SearchPath, flag.Args()[0])
	}

	if len(flag.Args()) > 1 {
		o.SearchGrep = append(o.SearchGrep, flag.Args()[1:]...)
	}
}

func (o *options) prepareContextLines(isTTY bool) {
	if !isTTY {
		o.ContextLines = 0
		o.AfterContextLines = 0
		o.BeforeContextLines = 0

		o.extra.actualAfterContextLines = 0
		o.extra.actualBeforeContextLines = 0

		o.extra.withAfterContextLines = false
		o.extra.withBeforeContextLines = false

		return
	}

	if o.AfterContextLines > 0 {
		o.extra.actualAfterContextLines = o.AfterContextLines
	} else if o.ContextLines > 0 {
		o.extra.actualAfterContextLines = o.ContextLines
	}

	o.extra.withAfterContextLines = o.ContextLines > 0 || o.AfterContextLines > 0

	if o.BeforeContextLines > 0 {
		o.extra.actualBeforeContextLines = o.BeforeContextLines
	} else if o.ContextLines > 0 {
		o.extra.actualBeforeContextLines = o.ContextLines
	}

	o.extra.withBeforeContextLines = o.ContextLines > 0 || o.BeforeContextLines > 0
}

func (o *options) prepareFromENV() {
	if os.Getenv(XFG_NO_COLOR_ENV_KEY) != "" {
		o.NoColor = true
	}
}

func (o *options) prepareAliases() {
	if o.Unrestricted {
		o.SearchAll = true
	}
}

func (o *options) prepareRuntimeFlags() {
	if len(o.SearchGrep) > 0 || len(o.SearchGrepRe) > 0 {
		o.extra.onlyMatchContent = true
	}
}

func (o *options) validateOptions() error {
	if err := validateStartPath(o.SearchStart); err != nil {
		return err
	}

	if len(o.Lang) > 0 {
		if err := validateLanguageCondition(o.Lang); err != nil {
			return err
		}
	}

	return nil
}

func showLangList() string {
	m := xfglangxt.List()
	languages := make([]string, 0, len(m))
	for k := range m {
		languages = append(languages, k)
	}
	sort.Strings(languages)

	out := ""
	for _, lang := range languages {
		out = out + lang + ": " + strings.Join(m[lang], ", ") + "\n"
	}

	return strings.TrimRight(out, "\n")
}

func versionDetails() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	compiler := runtime.Version()

	return fmt.Sprintf(
		"Version %s - %s.%s (compiled:%s, %s)",
		getVersion(),
		goos,
		goarch,
		compiler,
		installFrom,
	)
}

func getVersion() string {
	if version != "" {
		return version
	}
	i, ok := debug.ReadBuildInfo()
	if !ok {
		return "Unknown"
	}

	return i.Main.Version
}

func (cli *runner) putErr(message ...interface{}) {
	fmt.Fprintln(cli.err, message...)
}

func (cli *runner) putUsage() {
	cli.putErr(fmt.Sprintf("Usage: %s [SEARCH_PATH_KEYWORD] [SEARCH_CONTENT_KEYWORD] [OPTIONS]", cmdName))
}

func (cli *runner) putHelp(message string) {
	cli.putErr(message)
	cli.putUsage()
	cli.putErr("Options:")
	flag.PrintDefaults()
	funcExit(exitOK)
}
