package main

import (
	"os"
	"strings"
)

const supportTypes = "directory (d), symlink (l), executable (x), empty (e), socket (s), pipe (p), block-device (b), char-device (c)"

var message = map[string]map[string]string{
	"help_Stats": {
		"en": "Print runtime stats after searching result",
		"ja": "検索結果の後に、実行した処理の統計を表示します",
	},
	"help_SearchPath": {
		"en": "A string to find paths",
		"ja": "パスを検索するためのワード",
	},
	"help_SearchGrep": {
		"en": "A string to search for contents",
		"ja": "コンテンツを検索するためのワード",
	},
	"help_SearchStart": {
		"en": "A location to start searching",
		"ja": "検索を開始するディレクトリパス",
	},
	"help_IgnoreCase": {
		"en": "Ignore case distinctions to search. Also affects keywords of ignore option",
		"ja": "検索ワードの大文字小文字を区別しない。--ignore オプションでも有効化される",
	},
	"help_SearchPathRe": {
		"en": "A string to find paths by regular expressions (RE2)",
		"ja": "パスを検索するための正規表現 (RE2)",
	},
	"help_SearchGrepRe": {
		"en": "A string to grep contents by regular expressions (RE2)",
		"ja": "コンテンツを検索するための正規表現 (RE2)",
	},
	"help_NotWordBoundary": {
		"en": "Not care about word boundary to match by regexp",
		"ja": "正規表現でマッチする際に文字境界を無視してマッチするようにする",
	},
	"help_ContextLines": {
		"en": "Show several lines before and after the matched one",
		"ja": "マッチした行の前後 n 行も表示する",
	},
	"help_AfterContextLines": {
		"en": "Show several lines after the matched one. Override context option",
		"ja": "マッチした行以降 n 行を表示する",
	},
	"help_BeforeContextLines": {
		"en": "Show several lines before the matched one. Override context option",
		"ja": "マッチした行以前 n 行を表示する",
	},
	"help_Hidden": {
		"en": "Enable to search hidden files",
		"ja": "名前がドットではじまるファイルやディレクトリも検索対象にする",
	},
	"help_SkipGitIgnore": {
		"en": "Search files and directories even if a path matches a line of .gitignore",
		"ja": ".gitignore にマッチしたファイルやディレクトリも検索対象とする",
	},
	"help_SkipXfgIgnore": {
		"en": "Search files and directories even if a path matches a line of .xfgignore",
		"ja": ".xfgignore にマッチしたファイルやディレクトリも検索対象とする",
	},
	"help_NoDefaultSkip": {
		"en": "Not skip .git, .gitkeep, .gitkeep, .svn, node_modules, vendor, *.min.js and *.mmin.css",
		"ja": "次のファイルやディレクトリも検索対象とする .git, .gitkeep, .gitkeep, .svn, node_modules, vendor, *.min.js and *.mmin.css",
	},
	"help_SearchDefaultSkipStuff": {
		"en": "Search for hidden stuff and default skip files and directories)",
		"ja": "名前がドットではじまるものや、デフォルトで検索対象にならないファイルやディレクトリを検索対象にする",
	},
	"help_SearchAll": {
		"en": "Search all files and directories except specific ignoring files and directories",
		"ja": "すべてのファイルとディレクトリを検索対象とする",
	},
	"help_Unrestricted": {
		"en": "The alias of --search-all",
		"ja": "--search-all のエイリアス",
	},
	"help_Ignore": {
		"en": "Ignore path to pick up even with '--search-all'",
		"ja": "--search-all よりも優先してパスの検索から除外するワード",
	},
	"help_SearchOnlyName": {
		"en": "Search to only name instead whole path string",
		"ja": "パス全体ではなく、ファイルまたはディレクトリの名前だけを検索対象とする",
	},
	"help_Type": {
		"en": "Filter by file type: " + supportTypes,
		"ja": "ファイルタイプでフィルタする " + supportTypes,
	},
	"help_Ext": {
		"en": "Only search files matching file extension",
		"ja": "ファイル拡張子がマッチしたものだけ検索する",
	},
	"help_Lang": {
		"en": "Only search files matching language. --lang-list prints all support languages",
		"ja": "プログラミング言語を指定して検索する。--lang-list でサポートしている言語が一覧できる",
	},
	"help_flagLangList": {
		"en": "Show all supported file extensions for each language",
		"ja": "--lang で指定できる言語の一覧",
	},
	"help_Abs": {
		"en": "Show absolute paths",
		"ja": "絶対パスで表示する",
	},
	"help_ShowMatchCount": {
		"en": "Show a count of matching lines instead of contents",
		"ja": "マッチした行コンテンツを表示する代わりにマッチした回数を表示する",
	},
	"help_MaxMatchCount": {
		"en": "Stop reading a file after NUM matching lines",
		"ja": "1ファイルにつき、n回以上のマッチはスキップする",
	},
	"help_MaxColumns": {
		"en": "Do not print lines longer than this limit",
		"ja": "1行の長さが指定した長さを超える場合、表示しない",
	},
	"help_FilesWithMatches": {
		"en": "Print only the paths with at least one match",
		"ja": "マッチするファイルパスのみを表示する。ディレクトリパスやマッチしたコンテンツ自体は表示しない",
	},
	"help_Null": {
		"en": "Separate the filenames with \\0, rather than \\n",
		"ja": "結果表示でファイル群を \\n の代わりに \\0 で分割する",
	},
	"help_NoColor": {
		"en": "Disable colors for an output",
		"ja": "結果表示に色を付けない",
	},
	"help_ColorPathBase": {
		"en": "Color name for a path",
		"ja": "パスの色",
	},
	"help_ColorPath": {
		"en": "Color name to highlight keywords in a path",
		"ja": "パスの中でマッチした部分をハイライトする色",
	},
	"help_ColorContent": {
		"en": "Color name to highlight keywords in a content line",
		"ja": "コンテンツの中でマッチした部分をハイライトする色",
	},
	"help_GroupSeparator": {
		"en": "Print this string instead of '--' between groups of lines",
		"ja": "コンテンツの行のグループを分ける文字列",
	},
	"help_NoGroupSeparator": {
		"en": "Do not print a separator between groups of lines",
		"ja": "コンテンツの行のグループを分ける文字列を表示しない",
	},
	"help_Indent": {
		"en": "Indent string for the top of each line",
		"ja": "コンテンツ行のインデント",
	},
	"help_NoIndent": {
		"en": "Do not print an indent string",
		"ja": "コンテンツ行をインデントしない",
	},
	"help_IgnorePermissionError": {
		"en": "Do not print warnings of file permission error",
		"ja": "ファイル権限エラーを無視して、警告表示をしない",
	},
	"help_XfgIgnoreFile": {
		"en": ".xfgignore file path if you have it except XDG base directory or HOME directory",
		"ja": ".xfgignore ファイルのパス",
	},
	"help_NoPager": {
		"en": "Do not invoke with the Pager",
		"ja": "ページャーを無効にする",
	},
	"help_Quiet": {
		"en": "Do not write anything to standard output. Exit immediately with zero status if any match is found",
		"ja": "検索結果を標準出力に何もださず、マッチしたら exit status をゼロにする",
	},
	"help_Help": {
		"en": "Show help (This message) and exit",
		"ja": "ヘルプ（このメッセージ）を表示する",
	},
	"help_Version": {
		"en": "Show version and build command info and exit",
		"ja": "バージョンとビルドの情報を表示する",
	},
}

func isJa() bool {
	return strings.HasPrefix(strings.ToLower(os.Getenv("LANG")), "ja_jp.utf")
}

func getMessage(key string) string {
	lang := "en"
	if isJa() {
		lang = "ja"
	}

	if list, ok := message[key]; !ok {
		panic("not found message key " + key)
	} else if msg, ok := list[lang]; !ok {
		panic("not found message " + key + " in " + lang)
	} else {
		return msg
	}
}
