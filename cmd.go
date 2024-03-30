package main

import (
	"fmt"

	flag "github.com/spf13/pflag"
)

const (
	cmdName string = "xfg"
)

const (
	exitOK  int = 0
	exitErr int = 1
)

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
