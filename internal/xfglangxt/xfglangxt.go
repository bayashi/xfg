package xfglangxt

import "strings"

var langxt = map[string][]string{
	"ada":      {".ada", ".adb", ".ads"},
	"asciidoc": {".asc", ".asciidoc", ".adoc", ".ad"},
	"asm":      {".asm", ".s", ".S"},
	"asp":      {".asmx", ".aspx", ".aspx.cs", ".asax", ".asp", ".asa", ".ashx", ".ascx", ".ascx.cs", ".aspx.vb", ".ascx.vb"},
	"aspx":     {".ashx", ".ascx", ".asp", ".asa", ".aspx", ".asax", ".asmx"},
	"batch":    {".bat", ".cmd"},
	"bazel":    {".bazel", ".bzl", ".BUILD", ".bazelrc", "BUILD", "MODULE.bazel", "WORKSPACE", "WORKSPACE.bazel"},
	"bitbake":  {".bb", ".bbappend", ".bbclass", ".inc", ".conf"},
	"cc":       {".xs", ".c", ".h"},
	"cfmx":     {".cfc", ".cfm", ".cfml"},
	"clojure":  {".edn", ".cljc", ".cljx", ".clj", ".cljs"},
	"coffee":   {".coffee", ".cjsx"},
	"coq":      {".coq", ".g", ".v"},
	"cpp":      {".tpp", ".m", ".hpp", ".H", ".hxx", ".C", ".cxx", ".hh", ".h", ".cpp", ".cc"},
	"css":      {".css", "sass", ".scss"},
	"cython":   {".pxi", ".pyx", ".pxd"},
	"delphi":   {".bdsproj", ".pas", ".int", ".dpr", ".dproj", ".dfm", ".nfm", ".dof", ".dpk", ".groupproj", ".bdsgroup"},
	"ebuild":   {".ebuild", ".eclass"},
	"elixir":   {".ex", ".eex", ".exs"},
	"erlang":   {".erl", ".hrl"},
	"fortran":  {".F90", ".f95", ".ftn", ".fpp", ".f77", ".f90", ".f", ".F", ".f03", ".for", ".FPP"},
	"fsharp":   {".fsx", ".fs", ".fsi"},
	"gettext":  {".mo", ".po", ".pot"},
	"glsl":     {".vert", ".tesc", ".tese", ".geom", ".frag", ".comp"},
	"groovy":   {".gradle", ".gpp", ".grunit", ".groovy", ".gtmpl"},
	"haskell":  {".hs", ".hsig", ".lhs"},
	"html":     {".shtml", ".xhtml", ".htm", ".html"},
	"idris":    {".lidr", ".idr", ".ipkg"},
	"java":     {".java", ".properties"},
	"js":       {".es6", ".js", ".jsx", ".vue"},
	"jsp":      {".jsp", ".jspx", ".tagf", ".jhtm", ".jhtml", ".jspf", ".tag"},
	"make":     {".Makefiles", ".mk", ".mak"},
	"markdown": {".mdwn", ".mkdn", ".markdown", ".mdown", ".mkd", ".md"},
	"mason":    {".mas", ".mhtml", ".mpl", ".mtxt"},
	"md":       {".mkd", ".md", ".markdown", ".mdown", ".mdwn", ".mkdn"},
	"ocaml":    {".ml", ".mli", ".mll", ".mly"},
	"parrot":   {".pir", ".pasm", ".pmc", ".ops", ".pod", ".pg", ".tg"},
	"perl":     {".pl", ".pm", ".t", ".pod", ".PL"},
	"php":      {".php3", ".php4", ".php5", ".phtml", ".php", ".phpt"},
	"plone":    {".metadata", ".cpy", ".zcml", ".py", ".xml", ".pt", ".cpt"},
	"racket":   {".scm", ".rkt", ".ss"},
	"r":        {".r", ".R", ".Rtex", ".Rrst", ".Rmd", ".Rnw"},
	"ruby":     {".rjs", ".rxml", ".erb", ".rake", ".spec", ".rb", ".rhtml"},
	"shell":    {".csh", ".tcsh", ".ksh", ".zsh", ".sh", ".bash", ".fish"},
	"sml":      {".sml", ".fun", ".mlb", ".sig"},
	"tcl":      {".tcl", ".itcl", ".itk"},
	"tex":      {".sty", ".tex", ".cls"},
	"tt":       {".tt", ".tt2", ".ttml"},
	"ts":       {".ts", ".tsx"},
	"vala":     {".vala", ".vapi"},
	"vb":       {".bas", ".cls", ".vb", ".resx", ".frm", ".ctl"},
	"velocity": {".vsl", ".vm", ".vtl"},
	"verilog":  {".sv", ".svh", ".v", ".vh"},
	"vhdl":     {".vhd", ".vhdl"},
	"xml":      {".xsd", ".ent", ".xsl", ".xslt", ".tld", ".plist", ".wsdl", ".xml", ".dtd"},
	"yaml":     {".yaml", ".yml"},
	"yml":      {".yml", ".yaml"},
	"zeek":     {".bif", ".zeek", ".bro"},
}

func List() map[string][]string {
	return langxt
}

func IsSupported(lang string) bool {
	_, ok := langxt[strings.ToLower(lang)]
	return ok
}

func Get(lang string) []string {
	xt, ok := langxt[strings.ToLower(lang)]
	if !ok {
		return nil
	}

	return xt
}

func IsLangFile(lang string, filename string) bool {
	xt := Get(lang)
	if xt == nil {
		return false
	}

	for _, l := range xt {
		if strings.HasSuffix(filename, l) {
			return true
		}
	}

	return false
}
