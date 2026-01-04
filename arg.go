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
	defaultMaxDepth       uint32 = 255

	streamResultChanBufferSize int = 100
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
	SearchStart []string `toml:"start"`

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

	IgnoreCase             bool `toml:"ignore-case"`
	KeepResultOrder        bool `toml:"keep-result-order"`
	NoColor                bool `toml:"no-color"`
	Abs                    bool `toml:"abs"`
	ShowMatchCount         bool `toml:"count"`
	NoGroupSeparator       bool `toml:"no-group-separator"`
	NoIndent               bool `toml:"no-indent"`
	Hidden                 bool `toml:"hidden"`
	SkipGitIgnore          bool `toml:"skip-gitignore"`
	SkipXfgIgnore          bool `toml:"skip-xfgignore"`
	NoDefaultSkip          bool `toml:"no-default-skip"`
	SearchDefaultSkipStuff bool `toml:"search-default-skip-stuff"`
	SearchAll              bool `toml:"search-all"`
	Unrestricted           bool `toml:"unrestricted"`
	NoPager                bool `toml:"no-pager"`
	Quiet                  bool `toml:"quiet"`
	FilesWithMatches       bool `toml:"files-with-matches"`
	Null                   bool `toml:"null"`
	Stats                  bool `toml:"stats"`
	SearchOnlyName         bool `toml:"search-only-name"`
	NotWordBoundary        bool `toml:"not-word-boundary"`
	IgnorePermissionError  bool `toml:"ignore-permission-error"`

	flagLangList bool

	ContextLines uint32 `toml:"context"`

	AfterContextLines  uint32 `toml:"after-context"`
	BeforeContextLines uint32 `toml:"before-context"`

	MaxMatchCount uint32 `toml:"max-count"`
	MaxColumns    uint32 `toml:"max-columns"`
	MaxDepth      uint32 `toml:"max-depth"`

	extra optionsExtra
}

func (o *options) falgs(d *options) {
	flag.StringArrayVarP(&o.SearchPath, "path", "p", d.SearchPath, getMessage("help_SearchPath"))
	flag.StringArrayVarP(&o.SearchGrep, "grep", "g", d.SearchGrep, getMessage("help_SearchGrep"))
	flag.StringArrayVarP(&o.SearchStart, "start", "s", d.SearchStart, getMessage("help_SearchStart"))

	flag.BoolVarP(&o.IgnoreCase, "ignore-case", "i", d.IgnoreCase, getMessage("help_IgnoreCase"))
	flag.BoolVarP(&o.KeepResultOrder, "keep-result-order", "", d.KeepResultOrder, getMessage("help_KeepResultOrder"))

	flag.StringArrayVarP(&o.SearchPathRe, "path-regexp", "P", d.SearchPathRe, getMessage("help_SearchPathRe"))
	flag.StringArrayVarP(&o.SearchGrepRe, "grep-regexp", "G", d.SearchGrepRe, getMessage("help_SearchGrepRe"))
	flag.BoolVarP(&o.NotWordBoundary, "not-word-boundary", "M", d.NotWordBoundary, getMessage("help_NotWordBoundary"))

	flag.Uint32VarP(&o.ContextLines, "context", "C", d.ContextLines, getMessage("help_ContextLines"))
	flag.Uint32VarP(&o.AfterContextLines, "after-context", "A", d.AfterContextLines, getMessage("help_AfterContextLines"))
	flag.Uint32VarP(&o.BeforeContextLines, "before-context", "B", d.BeforeContextLines, getMessage("help_BeforeContextLines"))

	flag.BoolVarP(&o.Hidden, "hidden", ".", d.Hidden, getMessage("help_Hidden"))
	flag.BoolVarP(&o.SkipGitIgnore, "skip-gitignore", "", d.SkipGitIgnore, getMessage("help_SkipGitIgnore"))
	flag.BoolVarP(&o.SkipXfgIgnore, "skip-xfgignore", "", d.SkipXfgIgnore, getMessage("help_SkipXfgIgnore"))
	flag.BoolVarP(&o.NoDefaultSkip, "no-default-skip", "", d.NoDefaultSkip, getMessage("help_NoDefaultSkip"))
	flag.BoolVarP(&o.SearchDefaultSkipStuff, "search-default-skip-stuff", "n", d.SearchDefaultSkipStuff, getMessage("help_SearchDefaultSkipStuff"))
	flag.BoolVarP(&o.SearchAll, "search-all", "a", d.SearchAll, getMessage("help_SearchAll"))
	flag.BoolVarP(&o.Unrestricted, "unrestricted", "u", d.Unrestricted, getMessage("help_Unrestricted"))
	flag.StringArrayVarP(&o.Ignore, "ignore", "", d.Ignore, getMessage("help_Ignore"))
	flag.BoolVarP(&o.SearchOnlyName, "search-only-name", "f", d.SearchOnlyName, getMessage("help_SearchOnlyName"))

	flag.StringVarP(&o.Type, "type", "t", d.Type, getMessage("help_Type"))
	flag.StringArrayVarP(&o.Ext, "ext", "", d.Ext, getMessage("help_Ext"))
	flag.StringArrayVarP(&o.Lang, "lang", "", d.Lang, getMessage("help_Lang"))
	flag.BoolVarP(&o.flagLangList, "lang-list", "", false, getMessage("help_flagLangList"))

	flag.BoolVarP(&o.Abs, "abs", "", d.Abs, getMessage("help_Abs"))
	flag.BoolVarP(&o.ShowMatchCount, "count", "c", d.ShowMatchCount, getMessage("help_ShowMatchCount"))
	flag.Uint32VarP(&o.MaxMatchCount, "max-count", "m", d.MaxMatchCount, getMessage("help_MaxMatchCount"))
	flag.Uint32VarP(&o.MaxColumns, "max-columns", "", d.MaxColumns, getMessage("help_MaxColumns"))
	flag.Uint32VarP(&o.MaxDepth, "max-depth", "", d.MaxDepth, getMessage("help_MaxDepth"))
	flag.BoolVarP(&o.FilesWithMatches, "files-with-matches", "l", d.FilesWithMatches, getMessage("help_FilesWithMatches"))
	flag.BoolVarP(&o.Null, "null", "0", d.Null, getMessage("help_Null"))

	flag.BoolVarP(&o.NoColor, "no-color", "", d.NoColor, getMessage("help_NoColor"))
	flag.StringVarP(&o.ColorPathBase, "color-path-base", "", d.ColorPathBase, getMessage("help_ColorPathBase"))
	flag.StringVarP(&o.ColorPath, "color-path", "", d.ColorPath, getMessage("help_ColorPath"))
	flag.StringVarP(&o.ColorContent, "color-content", "", d.ColorContent, getMessage("help_ColorContent"))

	flag.StringVarP(&o.GroupSeparator, "group-separator", "", d.GroupSeparator, getMessage("help_GroupSeparator"))
	flag.BoolVarP(&o.NoGroupSeparator, "no-group-separator", "", d.NoGroupSeparator, getMessage("help_NoGroupSeparator"))

	flag.StringVarP(&o.Indent, "indent", "", d.Indent, getMessage("help_Indent"))
	flag.BoolVarP(&o.NoIndent, "no-indent", "", d.NoIndent, getMessage("help_NoIndent"))

	flag.BoolVarP(&o.IgnorePermissionError, "ignore-permission-error", "", d.IgnorePermissionError, getMessage("help_IgnorePermissionError"))

	flag.StringVarP(&o.XfgIgnoreFile, "xfgignore-file", "", d.XfgIgnoreFile, getMessage("help_XfgIgnoreFile"))

	flag.BoolVarP(&o.NoPager, "no-pager", "", d.NoPager, getMessage("help_NoPager"))
	flag.BoolVarP(&o.Quiet, "quiet", "q", d.Quiet, getMessage("help_Quiet"))
	flag.BoolVarP(&o.Stats, "stats", "", d.Stats, getMessage("help_Stats"))
}

func (cli *runner) parseArgs(d *options) *options {
	o := &options{}
	o.extra.runWithNoArg = len(os.Args) == 1
	o.falgs(d)

	flag.CommandLine.SetOutput(cli.err)
	flag.CommandLine.SortFlags = false
	var flagHelp, flagVersion bool
	flag.BoolVarP(&flagHelp, "help", "h", false, getMessage("help_Help"))
	flag.BoolVarP(&flagVersion, "version", "v", false, getMessage("help_Version"))
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
