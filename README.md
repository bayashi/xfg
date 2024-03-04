# xfg

Do `find` paths by a keyword, and also search for contents like a `grep` in one command, gracefully.

<a href="https://github.com/bayashi/xfg/actions" title="xfg CI"><img src="https://github.com/bayashi/xfg/workflows/main/badge.svg" alt="xfg CI"></a>
<a href="https://goreportcard.com/report/github.com/bayashi/xfg" title="xfg report card" target="_blank"><img src="https://goreportcard.com/badge/github.com/bayashi/xfg" alt="xfg report card"></a>
<a href="https://pkg.go.dev/github.com/bayashi/xfg" title="Go xfg package reference" target="_blank"><img src="https://pkg.go.dev/badge/github.com/bayashi/xfg.svg" alt="Go Reference: xfg"></a>

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
* `.git` directory is ignored

## Installation

```cmd
go install github.com/bayashi/xfg@latest
```

## License

MIT License

## Author

Dai Okabayashi: https://github.com/bayashi
