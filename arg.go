package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	flag "github.com/spf13/pflag"
)

const (
	cmdName string = "xfg"

	XFG_RC_FILE string = ".xfgrc"

	errNeedToSetPath = "Err: You should specify a directory path `--path`"

	defaultGroupSeparator = "--"
	defaultIndent         = " "
)

var (
	version     = ""
	installFrom = "Source"
)

type options struct {
	SearchPath  []string `toml:"path"`
	SearchGrep  []string `toml:"grep"`
	SearchStart string   `toml:"start"`

	GroupSeparator string `toml:"gourp-separator"`
	Indent         string `toml:"indent"`
	ColorPath      string `toml:"color-path"`
	ColorContent   string `toml:"color-conetnt"`
	XfgIgnoreFile  string `toml:"xfgignore-file"`

	Ignore []string `toml:"ignore"`

	IgnoreCase       bool `toml:"ignore-case"`
	Relax            bool `toml:"relax"`
	NoColor          bool `toml:"no-color"`
	Abs              bool `toml:"abs"`
	ShowMatchCount   bool `toml:"count"`
	OnlyMatchContent bool `toml:"only-match"`
	NoGroupSeparator bool `toml:"no-group-separator"`
	NoIndent         bool `toml:"no-indent"`
	Hidden           bool `toml:"hidden"`
	SkipGitIgnore    bool `toml:"skip-gitignore"`
	SkipXfgIgnore    bool `toml:"skip-xfgignore"`
	SearchAll        bool `toml:"search-all"`
	NoPager          bool `toml:"no-pager"`
	Quiet            bool `toml:"quiet"`

	ContextLines uint32 `toml:"context"`

	AfterContextLines  uint32 `toml:"after-context"`
	BeforeContextLines uint32 `toml:"before-context"`

	MaxMatchCount uint32 `toml:"max-count"`
	MaxColumns    uint32 `toml:"max-columns"`

	// runtime options
	actualAfterContextLines  uint32
	actualBeforeContextLines uint32
	withAfterContextLines    bool
	withBeforeContextLines   bool
}

func (cli *runner) parseArgs(d *options) *options {
	noArgs := len(os.Args) == 1

	o := &options{}

	flag.CommandLine.SetOutput(cli.err)
	flag.CommandLine.SortFlags = false

	var flagHelp bool
	var flagVersion bool
	flag.StringArrayVarP(&o.SearchPath, "path", "p", d.SearchPath, "A string to find paths")
	flag.StringArrayVarP(&o.SearchGrep, "grep", "g", d.SearchGrep, "A string to search for contents")
	flag.StringVarP(&o.SearchStart, "start", "s", d.SearchStart, "A location to start searching")

	flag.BoolVarP(&o.IgnoreCase, "ignore-case", "i", d.IgnoreCase, "Ignore case distinctions to search. Also affects keywords of ignore option")
	flag.BoolVarP(&o.OnlyMatchContent, "only-match", "o", d.OnlyMatchContent, "Show paths only matched contents")

	flag.Uint32VarP(&o.ContextLines, "context", "C", d.ContextLines, "Show several lines before and after the matched one")
	flag.Uint32VarP(&o.AfterContextLines, "after-context", "A", d.AfterContextLines, "Show several lines after the matched one. Override context option")
	flag.Uint32VarP(&o.BeforeContextLines, "before-context", "B", d.BeforeContextLines, "Show several lines before the matched one. Override context option")

	flag.BoolVarP(&o.Hidden, "hidden", ".", d.Hidden, "Enable to search hidden files")
	flag.BoolVarP(&o.SkipGitIgnore, "skip-gitignore", "", d.SkipGitIgnore, "Search files and directories even if a path matches a line of .gitignore")
	flag.BoolVarP(&o.SkipXfgIgnore, "skip-xfgignore", "", d.SkipXfgIgnore, "Search files and directories even if a path matches a line of .xfgignore")
	flag.BoolVarP(&o.SearchAll, "search-all", "", d.SearchAll, "Search all files and directories except specific ignoring files and directories")
	flag.StringArrayVarP(&o.Ignore, "ignore", "", d.Ignore, "Ignore path to pick up even with '--search-all'")

	flag.BoolVarP(&o.Relax, "relax", "", d.Relax, "Insert blank space between contents for relaxing view")
	flag.BoolVarP(&o.Abs, "abs", "", d.Abs, "Show absolute paths")
	flag.BoolVarP(&o.ShowMatchCount, "count", "c", d.ShowMatchCount, "Show a count of matching lines instead of contents")
	flag.Uint32VarP(&o.MaxMatchCount, "max-count", "m", d.MaxMatchCount, "Stop reading a file after NUM matching lines")
	flag.Uint32VarP(&o.MaxColumns, "max-columns", "", d.MaxColumns, "Do not print lines longer than this limit")

	flag.BoolVarP(&o.NoColor, "no-color", "", d.NoColor, "Disable colors for an output")
	flag.StringVarP(&o.ColorPath, "color-path", "", d.ColorPath, "Color name to highlight keywords in a path")
	flag.StringVarP(&o.ColorContent, "color-content", "", d.ColorContent, "Color name to highlight keywords in a content line")

	flag.StringVarP(&o.GroupSeparator, "group-separator", "", d.GroupSeparator, "Print this string instead of '--' between groups of lines")
	flag.BoolVarP(&o.NoGroupSeparator, "no-group-separator", "", d.NoGroupSeparator, "Do not print a separator between groups of lines")

	flag.StringVarP(&o.Indent, "indent", "", d.Indent, "Indent string for the top of each line")
	flag.BoolVarP(&o.NoIndent, "no-indent", "", d.NoIndent, "Do not print an indent string")

	flag.StringVarP(&o.XfgIgnoreFile, "xfgignore-file", "", d.XfgIgnoreFile, ".xfgignore file path if you have it except XDG base directory or HOME directory")

	flag.BoolVarP(&o.NoPager, "no-pager", "", d.NoPager, "Do not invoke with the Pager")
	flag.BoolVarP(&o.Quiet, "quiet", "q", d.Quiet, "Do not write anything to standard output. Exit immediately with zero status if any match is found")

	flag.BoolVarP(&flagHelp, "help", "h", false, "Show help (This message) and exit")
	flag.BoolVarP(&flagVersion, "version", "v", false, "Show version and build command info and exit")

	flag.Parse()

	if noArgs || flagHelp {
		cli.putHelp(fmt.Sprintf("Version %s", getVersion()))
	} else if flagVersion {
		cli.putErr(versionDetails())
		funcExit(exitOK)
	} else if len(o.SearchPath) == 0 && len(flag.Args()) == 0 {
		cli.putHelp(errNeedToSetPath)
	}

	o.targetPathFromArgs()

	if len(o.SearchGrep) > 0 {
		o.OnlyMatchContent = true
	}

	return o
}

func (o *options) targetPathFromArgs() {
	if len(flag.Args()) > 0 && flag.Args()[0] != "" {
		o.SearchPath = append(o.SearchPath, flag.Args()[0])
	}

	if len(flag.Args()) == 2 && flag.Args()[1] != "" {
		o.SearchGrep = append(o.SearchGrep, flag.Args()[1])
	}
}

func (o *options) prepareContextLines(isTTY bool) {
	if !isTTY {
		o.ContextLines = 0
		o.AfterContextLines = 0
		o.BeforeContextLines = 0

		o.actualAfterContextLines = 0
		o.actualBeforeContextLines = 0

		o.withAfterContextLines = false
		o.withBeforeContextLines = false

		return
	}

	if o.AfterContextLines > 0 {
		o.actualAfterContextLines = o.AfterContextLines
	} else if o.ContextLines > 0 {
		o.actualAfterContextLines = o.ContextLines
	}

	o.withAfterContextLines = o.ContextLines > 0 || o.AfterContextLines > 0

	if o.BeforeContextLines > 0 {
		o.actualBeforeContextLines = o.BeforeContextLines
	} else if o.ContextLines > 0 {
		o.actualBeforeContextLines = o.ContextLines
	}

	o.withBeforeContextLines = o.ContextLines > 0 || o.BeforeContextLines > 0
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
