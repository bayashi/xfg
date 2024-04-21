# xfg

<a href="https://github.com/bayashi/xfg/actions" title="xfg CI"><img src="https://github.com/bayashi/xfg/workflows/main/badge.svg" alt="xfg CI"></a>
<a href="https://goreportcard.com/report/github.com/bayashi/xfg" title="xfg report card" target="_blank"><img src="https://goreportcard.com/badge/github.com/bayashi/xfg" alt="xfg report card"></a>
<a href="https://pkg.go.dev/github.com/bayashi/xfg" title="Go xfg package reference" target="_blank"><img src="https://pkg.go.dev/badge/github.com/bayashi/xfg.svg" alt="Go Reference: xfg"></a>

Find paths anyway, then search for contents also, naturally.

* Recursive search
* Search both paths and contents by multiple keywords
    * Possible to search by regexp also
* Ignores hidden files and directories
* Respect your `.gitignore`
    * Possible to use specific `.xfgignore` also
* Support `.xfgrc` file for your default options
* support to search specific language files by `--lang` option
    * `--lang perl` will search for files which extentions are ".pl", ".pm", ".t", ".pod" or ".PL"
    * `xfg --lang-list` to show all supported languages

There are so many features. You can check all options in below "Help Options" section.

## Usage of `xfg` command

Search for files and directories that include `service-b` in those path.

```sh
$ xfg service-b
```

Specific:

```sh
$ xfg --path service-b
```

By default, start searching from current directory. You can specify `--start` option for where to start.

Output:

```
$ xfg service-b
service-b
service-b/main.go
```

Search for files and directories that match the `service-b` in those path and extract content that matches the `bar`.

```sh
$ xfg service-b bar
```

Specific:

```
$ xfg --path service-b --grep bar
```

Output:

```
service-b/main.go
  4:    bar := 34
```

Note that the second argument and subsequent arguments are tereated as keywords to grep contents

```sh
$ xfg service-b bar baz
        |        |   |
        |        |   +--- To search for contents
        |        +------- To search for contents
        +---------------- To search for paths
```

Above command is equivalent:

```sh
$ xfg --path service-b --grep bar --grep baz
```

You can use multiple keywords to match for both `--path` and `--grep` like below

```
$ xfg --path foo --path bar --grep baz --grep qux
```

These keywords are treated as AND condition for each.

## Notes

* Not follow symbolic links
* Skip binary files

## Regexp search

xfg can search for paths and contents by regexp.

`-P` is to search for paths

```sh
$ xfg -P "service-[a-g]"
```

`-G` is to search for contents

```sh
$ xfg -G "X_[A-Z]+_[A-Z]+"
```

Regexp keywords you input respect word boundaries by default. You can use `--not-word-boundary` option to trun it off.

### Ignore rules

* Ignored `*.min.js` or `*.min.css` files by default
* Ignored `.git`, `.svn`, `node_modules` or `vendor` directories and files in them by default
* Files and directories whose names start with a dot are ignored by default
    * Use `-.` or `--hidden` option to search for hidden-files/directories
        * Still ignore `.git`, `.gitkeep`, `node_modules`, `vendor`, `.gitkeep`, `*.min.js` and `*.mmin.css`
* `-a` or `--search-all` option enables to search for all files and directories
    * You can ignore specific files and directories with `--ignore` option

### .gitignore file

* Support .gitignore file to ignore files and directories by default
    * Read `.gitignore` file from starting directory to search or HOME directory.
    * Use `--skip-gitignore` option to avoid reading `.gitignore` file.
* Support .xfgignore file to ignore files and directories as same as .gitignore by default
    * `.xfgignore` file should be located in [XDG Base Directory](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) or HOME directory
    * You can specify `.xfgignore` file path by `--xfgignore-file` option
    * Use `--skip-xfgignore` option to avoid reading `.xfgignore` file

## Help Options

```
  -p, --path stringArray          A string to find paths
  -g, --grep stringArray          A string to search for contents
  -s, --start string              A location to start searching (default ".")
  -i, --ignore-case               Ignore case distinctions to search. Also affects keywords of ignore option
  -P, --path-regexp stringArray   A string to find paths by regular expressions (RE2)
  -G, --grep-regexp stringArray   A string to grep contents by regular expressions (RE2)
      --not-word-boundary         Not care about word boundary to match by regexp
  -C, --context uint32            Show several lines before and after the matched one
  -A, --after-context uint32      Show several lines after the matched one. Override context option
  -B, --before-context uint32     Show several lines before the matched one. Override context option
  -., --hidden                    Enable to search hidden files
      --skip-gitignore            Search files and directories even if a path matches a line of .gitignore
      --skip-xfgignore            Search files and directories even if a path matches a line of .xfgignore
  -a, --search-all                Search all files and directories except specific ignoring files and directories
  -u, --unrestricted              The alias of --search-all
      --ignore stringArray        Ignore path to pick up even with '--search-all'
  -f, --search-only-name          Search to only name instead whole path string
      --ext stringArray           Only search files matching file extension
      --lang stringArray          Only search files matching language. --type-list prints all support languages
      --lang-list                 Show all supported file extensions for each language
      --abs                       Show absolute paths
  -c, --count                     Show a count of matching lines instead of contents
  -m, --max-count uint32          Stop reading a file after NUM matching lines
      --max-columns uint32        Do not print lines longer than this limit
  -l, --files-with-matches        Only print the names of matching files
  -0, --null                      Separate the filenames with \0, rather than \n
      --no-color                  Disable colors for an output
      --color-path-base string    Color name for a path (default "yellow")
      --color-path string         Color name to highlight keywords in a path (default "cyan")
      --color-content string      Color name to highlight keywords in a content line (default "red")
      --group-separator string    Print this string instead of '--' between groups of lines (default "--")
      --no-group-separator        Do not print a separator between groups of lines
      --indent string             Indent string for the top of each line (default " ")
      --no-indent                 Do not print an indent string
      --xfgignore-file string     .xfgignore file path if you have it except XDG base directory or HOME directory
      --no-pager                  Do not invoke with the Pager
  -q, --quiet                     Do not write anything to standard output. Exit immediately with zero status if any match is found
      --stats                     Print runtime stats after searching result
  -h, --help                      Show help (This message) and exit
  -v, --version                   Show version and build command info and exit
```

## Default Options

You can use default options by `.xfgrc` file. It's TOML file even without `.toml` extention, anyway.

```
color-path = "blue"
ignore = [
    ".vscode",
    ".idea",
]
```

`.xfgrc` file should be in [XDG Base Directory](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) or HOME directory. Or, you can specify a file path by ENV key: `XFG_RC_FILE_PATH` as you like.

```
export XFG_RC_FILE_PATH="/path/to/your_rc_file.toml"
```

## Highlight limitation

xfg can highlight matched keywords in results. But the highlight feature is not perfect yet.

The `xfg` adds colors for highlight keywords for paths. However, if you use both (`-P`, `--path-regexp`) option and (`-p`, `--path`) option at the same time, and when both conditions are matching with same peace of string, then the ONLY (`-P`, `--path-regexp`) condition can highlight string so far. This limitation is same as grep contents condition. Regexp conditions are strong to be highlighted string than word match condition when both conditions matches with same pease of string.

Moreover, not yet highlighted extentions by `--lang` or `--ext` options even if it's matched.

## Installation

### Mac

```sh
brew tap bayashi/tap
brew install bayashi/tap/xfg
```

### Binary install

Download binary from here: https://github.com/bayashi/xfg/releases

### Go manual

If you have golang environment:

```cmd
go install github.com/bayashi/xfg@latest
```

## License

MIT License

## Author

Dai Okabayashi: https://github.com/bayashi
