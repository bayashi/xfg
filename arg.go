package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	flag "github.com/spf13/pflag"
)

const (
	exitOK  int = 0
	exitErr int = 1
)

const (
	cmdName string = "xfg"

	errNeedToSetPath = "Err: You should specify a directory path `--path`"

	defaultGroupSeparator = "--"
	defaultIndent         = " "
)

var (
	version     = ""
	installFrom = "Source"
)

type options struct {
	searchPath  []string
	searchGrep  []string
	searchStart string

	groupSeparator string
	indent         string
	colorPath      string
	colorContent   string

	ignore []string

	ignoreCase       bool
	relax            bool
	noColor          bool
	abs              bool
	showMatchCount   bool
	onlyMatchContent bool
	noGroupSeparator bool
	noIndent         bool
	hidden           bool
	skipGitIgnore    bool
	searchAll        bool
	noPager          bool
	quiet            bool

	contextLines uint32

	afterContextLines  uint32
	beforeContextLines uint32

	maxMatchCount uint32
	maxColumns    uint32

	// runtime options
	actualAfterContextLines  uint32
	actualBeforeContextLines uint32
	withAfterContextLines    bool
	withBeforeContextLines   bool
}

func (cli *runner) parseArgs() *options {
	noArgs := len(os.Args) == 1

	o := &options{}

	flag.CommandLine.SetOutput(cli.err)
	flag.CommandLine.SortFlags = false

	var flagHelp bool
	var flagVersion bool
	flag.StringArrayVarP(&o.searchPath, "path", "p", []string{}, "A string to find paths")
	flag.StringArrayVarP(&o.searchGrep, "grep", "g", []string{}, "A string to search for contents")
	flag.StringVarP(&o.searchStart, "start", "s", ".", "A location to start searching")

	flag.Uint32VarP(&o.afterContextLines, "after-context", "A", 0, "Show several lines after the matched one. Override context option")
	flag.Uint32VarP(&o.beforeContextLines, "before-context", "B", 0, "Show several lines before the matched one. Override context option")
	flag.Uint32VarP(&o.contextLines, "context", "C", 0, "Show several lines before and after the matched one")
	flag.Uint32VarP(&o.maxMatchCount, "max-count", "m", 0, "Stop reading a file after NUM matching lines")
	flag.Uint32VarP(&o.maxColumns, "max-columns", "", 0, "Do not print lines longer than this limit")

	flag.StringArrayVarP(&o.ignore, "ignore", "", []string{}, "Ignore path to pick up even with '--search-all'")
	flag.BoolVarP(&o.hidden, "hidden", ".", false, "Enable to search hidden files")
	flag.BoolVarP(&o.skipGitIgnore, "skip-git-ignore", "", false, "Search files and directories even if a path matches a line of .gitignore")
	flag.BoolVarP(&o.searchAll, "search-all", "", false, "Search all files and directories except specific ignoring files and directories")
	flag.BoolVarP(&o.ignoreCase, "ignore-case", "i", false, "Ignore case distinctions to search. Also affects keywords of ignore option")

	flag.BoolVarP(&o.noColor, "no-color", "", false, "Disable colors for an output")
	flag.BoolVarP(&o.relax, "relax", "", false, "Insert blank space between contents for relaxing view")
	flag.BoolVarP(&o.abs, "abs", "", false, "Show absolute paths")
	flag.BoolVarP(&o.showMatchCount, "count", "c", false, "Show a count of matching lines instead of contents")
	flag.BoolVarP(&o.onlyMatchContent, "only-match", "o", false, "Show paths only matched contents")
	flag.BoolVarP(&o.noGroupSeparator, "no-group-separator", "", false, "Do not print a separator between groups of lines")
	flag.BoolVarP(&o.noIndent, "no-indent", "", false, "Do not print an indent string")
	flag.BoolVarP(&o.noPager, "no-pager", "", false, "Do not invoke with the Pager")
	flag.BoolVarP(&o.quiet, "quiet", "q", false, "Do not write anything to standard output. Exit immediately with zero status if any match is found")

	flag.StringVarP(&o.groupSeparator, "group-separator", "", defaultGroupSeparator, "Print this string instead of '--' between groups of lines")
	flag.StringVarP(&o.indent, "indent", "", defaultIndent, "Indent string for the top of each line")
	flag.StringVarP(&o.colorPath, "color-path", "", "cyan", "Color name to highlight keywords in a path")
	flag.StringVarP(&o.colorContent, "color-content", "", "red", "Color name to highlight keywords in a content line")

	flag.BoolVarP(&flagHelp, "help", "h", false, "Show help (This message) and exit")
	flag.BoolVarP(&flagVersion, "version", "v", false, "Show version and build command info and exit")

	flag.Parse()

	if noArgs || flagHelp {
		cli.putHelp(fmt.Sprintf("Version %s", getVersion()))
	} else if flagVersion {
		cli.putErr(versionDetails())
		funcExit(exitOK)
	} else if len(o.searchPath) == 0 && len(flag.Args()) == 0 {
		cli.putHelp(errNeedToSetPath)
	}

	o.targetPathFromArgs()

	return o
}

func (o *options) targetPathFromArgs() {
	if len(flag.Args()) > 0 && flag.Args()[0] != "" {
		o.searchPath = append(o.searchPath, flag.Args()[0])
	}

	if len(flag.Args()) == 2 && flag.Args()[1] != "" {
		o.searchGrep = append(o.searchGrep, flag.Args()[1])
	}
}

func (o *options) prepareContextLines(isTTY bool) {
	if !isTTY {
		o.contextLines = 0
		o.afterContextLines = 0
		o.beforeContextLines = 0

		o.actualAfterContextLines = 0
		o.actualBeforeContextLines = 0

		o.withAfterContextLines = false
		o.withBeforeContextLines = false

		return
	}

	if o.afterContextLines > 0 {
		o.actualAfterContextLines = o.afterContextLines
	} else if o.contextLines > 0 {
		o.actualAfterContextLines = o.contextLines
	}

	o.withAfterContextLines = o.contextLines > 0 || o.afterContextLines > 0

	if o.beforeContextLines > 0 {
		o.actualBeforeContextLines = o.beforeContextLines
	} else if o.contextLines > 0 {
		o.actualBeforeContextLines = o.contextLines
	}

	o.withBeforeContextLines = o.contextLines > 0 || o.beforeContextLines > 0
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
