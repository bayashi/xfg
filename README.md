# xfg

Do `find` paths by a keyword, and also search for contents like a `grep` in one command.

<a href="https://github.com/bayashi/xfg/actions" title="xfg CI"><img src="https://github.com/bayashi/xfg/workflows/main/badge.svg" alt="xfg CI"></a>
<a href="https://goreportcard.com/report/github.com/bayashi/xfg" title="xfg report card" target="_blank"><img src="https://goreportcard.com/badge/github.com/bayashi/xfg" alt="xfg report card"></a>
<a href="https://pkg.go.dev/github.com/bayashi/xfg" title="Go xfg package reference" target="_blank"><img src="https://pkg.go.dev/badge/github.com/bayashi/xfg.svg" alt="Go Reference: xfg"></a>

## Usage of `xfg` command

Search for directories and files that include `service-a` in those path.

```go
$ xfg service-a
```

Search for directories and files that match the `service-a` in those path and extract content that matches the `memory`.

```go
$ xfg service-a --grep memory
```

`xfg` is the shorthand and enhancement of below one-liner.

```go
find . -type f | grep "service-id" | xargs grep "memory" -n
```

Name `xfg` came from Cross(**X**) **F**ind and **G**rep.

## Installation

```cmd
go install github.com/bayashi/xfg@latest
```

## License

MIT License

## Author

Dai Okabayashi: https://github.com/bayashi
