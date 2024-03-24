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

### Ignore rules

* Files and directories whose names start with a dot are ignored by default
    * Use `--hidden` option to search for hidden-files/directories
        * Still ignore `.git` and `.gitkeep`
* Ignored `*.min.js` file by default
* Support .gitignore file to ignore files and directories by default
    * Use `--skip-gitignore` option to enable `.gitignore` file
* Pick up all files and directories with `--search-all` option
    * You can ignore specific files and directories with `--ignore` option

Also See `--help` for more options.

## Installation

```cmd
go install github.com/bayashi/xfg@latest
```

## License

MIT License

## Author

Dai Okabayashi: https://github.com/bayashi
