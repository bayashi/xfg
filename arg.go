package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	flag "github.com/spf13/pflag"
)

const (
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
	onlyMatch        bool
	noGroupSeparator bool
	noIndent         bool
	hidden           bool
	skipGitIgnore    bool
	searchAll        bool

	contextLines  uint32
	maxMatchCount uint32
}

func (cli *runner) parseArgs() *options {
	noArgs := len(os.Args) == 1

	o := &options{}

	var flagHelp bool
	var flagVersion bool
	flag.StringArrayVarP(&o.searchPath, "path", "p", []string{}, "A string to find paths")
	flag.StringArrayVarP(&o.searchGrep, "grep", "g", []string{}, "A string to search for contents")
	flag.StringVarP(&o.searchStart, "start", "s", ".", "A location to start searching")

	flag.Uint32VarP(&o.contextLines, "context", "C", 0, "Show several lines before and after the matched one")
	flag.Uint32VarP(&o.maxMatchCount, "max-count", "m", 0, "Stop reading a file after NUM matching lines")

	flag.StringArrayVarP(&o.ignore, "ignore", "", []string{}, "Ignore path to pick up even with '--search-all'")
	flag.BoolVarP(&o.hidden, "hidden", "", false, "Enable to search hidden files")
	flag.BoolVarP(&o.skipGitIgnore, "skip-git-ignore", "", false, "Search files and directories even if a path matches a line of .gitignore")
	flag.BoolVarP(&o.searchAll, "search-all", "", false, "Search all files and directories except specific ignoring files and directories")
	flag.BoolVarP(&o.ignoreCase, "ignore-case", "i", false, "Ignore case distinctions to search. Also affects keywords of ignore option")

	flag.BoolVarP(&o.noColor, "no-color", "", false, "Disable colors for an output")
	flag.BoolVarP(&o.relax, "relax", "", false, "Insert blank space between contents for relaxing view")
	flag.BoolVarP(&o.abs, "abs", "", false, "Show absolute paths")
	flag.BoolVarP(&o.showMatchCount, "count", "c", false, "Show a count of matching lines instead of contents")
	flag.BoolVarP(&o.onlyMatch, "only-match", "o", false, "Show paths only matched contents")
	flag.BoolVarP(&o.noGroupSeparator, "no-group-separator", "", false, "Do not print a separator between groups of lines")
	flag.BoolVarP(&o.noIndent, "no-indent", "", false, "Do not print an indent string")

	flag.StringVarP(&o.groupSeparator, "group-separator", "", defaultGroupSeparator, "Print this string instead of '--' between groups of lines")
	flag.StringVarP(&o.indent, "indent", "", defaultIndent, "Indent string for the top of each line")
	flag.StringVarP(&o.colorPath, "color-path", "", "cyan", "Color name to highlight keywords in a path")
	flag.StringVarP(&o.colorContent, "color-content", "", "red", "Color name to highlight keywords in a content line")

	flag.BoolVarP(&flagHelp, "help", "h", false, "Show help (This message) and exit")
	flag.BoolVarP(&flagVersion, "version", "v", false, "Show version and build command info and exit")

	flag.CommandLine.SortFlags = false
	flag.Parse()

	if noArgs || flagHelp {
		cli.putHelp(fmt.Sprintf("Version %s", getVersion()))
	}

	if flagVersion {
		cli.putErr(versionDetails())
		os.Exit(exitOK)
	}

	o.targetPathFromArgs(cli)

	return o
}

func (o *options) targetPathFromArgs(cli *runner) {
	if len(flag.Args()) == 0 {
		cli.putHelp(errNeedToSetPath)
	}

	o.searchPath = append(o.searchPath, flag.Args()[0])

	if len(flag.Args()) == 2 {
		o.searchGrep = append(o.searchGrep, flag.Args()[1])
	}
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
