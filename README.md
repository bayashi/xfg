# xfg

<a href="https://github.com/bayashi/xfg/actions" title="xfg CI"><img src="https://github.com/bayashi/xfg/workflows/main/badge.svg" alt="xfg CI"></a>
<a href="https://goreportcard.com/report/github.com/bayashi/xfg" title="xfg report card" target="_blank"><img src="https://goreportcard.com/badge/github.com/bayashi/xfg" alt="xfg report card"></a>
<a href="https://pkg.go.dev/github.com/bayashi/xfg" title="Go xfg package reference" target="_blank"><img src="https://pkg.go.dev/badge/github.com/bayashi/xfg.svg" alt="Go Reference: xfg"></a>

Find paths anyway, then search for contents also

## Usage of `xfg` command

For example, there are directories and files like below:

```
.
├── service-a
│   └── main.go
├── service-b
│   └── main.go
└── service-c
    └── main.go
```

Search for directories and files that include `service-b` in those path.

```sh
$ xfg service-b
```

Specific:

```sh
$ xfg --path service-b
```

Output:

```
$ xfg service-b
service-b
service-b/main.go
```

Search for directories and files that match the `service-b` in those path and extract content that matches the `bar`.

```sh
$ xfg service-b bar
```

Specific:

```
$ xfg --path service-b --grep bar
```

Output:

```
service-b
service-b/main.go
  4:    bar := 34
```

## Notes

* Not follow symbolic links
* Skip binary files

### Ignore rules

* Files and directories whose names start with a dot are ignored by default
    * Use `--hidden` option to search for hidden-files/directories
        * Still ignore `.git` and `.gitkeep`
* Ignored `*.min.js` file by default
* Support .gitignore file to ignore files and directories by default
    * Use `--skip-gitignore` option to enable `.gitignore` file
* Pick up all files and directories with `--search-all` option
    * You can ignore specific files and directories with `--ignore` option

## Options

```
  -p, --path stringArray         A string to find paths
  -g, --grep stringArray         A string to search for contents
  -s, --start string             A location to start searching (default ".")
  -A, --after-context uint32     Show several lines after the matched one. Override context option
  -B, --before-context uint32    Show several lines before the matched one. Override context option
  -C, --context uint32           Show several lines before and after the matched one
  -m, --max-count uint32         Stop reading a file after NUM matching lines
      --ignore stringArray       Ignore path to pick up even with '--search-all'
  -., --hidden                   Enable to search hidden files
      --skip-git-ignore          Search files and directories even if a path matches a line of .gitignore
      --search-all               Search all files and directories except specific ignoring files and directories
  -i, --ignore-case              Ignore case distinctions to search. Also affects keywords of ignore option
      --no-color                 Disable colors for an output
      --relax                    Insert blank space between contents for relaxing view
      --abs                      Show absolute paths
  -c, --count                    Show a count of matching lines instead of contents
  -o, --only-match               Show paths only matched contents
      --no-group-separator       Do not print a separator between groups of lines
      --no-indent                Do not print an indent string
      --no-pager                 Do not invoke with the Pager
  -q, --quiet                    Do not write anything to standard output. Exit immediately with zero status if any match is found
      --group-separator string   Print this string instead of '--' between groups of lines (default "--")
      --indent string            Indent string for the top of each line (default " ")
      --color-path string        Color name to highlight keywords in a path (default "cyan")
      --color-content string     Color name to highlight keywords in a content line (default "red")
  -h, --help                     Show help (This message) and exit
  -v, --version                  Show version and build command info and exit
```

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
