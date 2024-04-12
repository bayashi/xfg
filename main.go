package main

import (
	"fmt"
	"io"
	"os"
)

const (
	exitOK  int = 0
	exitErr int = 1
)

var funcExit = func(code int) {
	os.Exit(code)
}

type runner struct {
	out      io.Writer
	err      io.Writer
	isTTY    bool
	exitCode int
	homeDir  string
}

func main() {
	cli := &runner{
		out:   os.Stdout,
		err:   os.Stderr,
		isTTY: isTTY(),
	}
	exitCode, message := cli.run()
	if exitCode != exitOK {
		cli.putErr(fmt.Sprintf("Err: %s", message))
	}
	funcExit(exitCode)
}

func (cli *runner) run() (int, string) {
	homeDir, err := homeDir()
	if err != nil {
		return exitErr, fmt.Sprintf("on detecting home directory : %s", err)
	}
	cli.homeDir = homeDir

	defaultOpt, err := readRC(cli.homeDir)
	if err != nil {
		return exitErr, fmt.Sprintf("on reading home directory : %s", err)
	}

	o := cli.parseArgs(defaultOpt)
	if !cli.isTTY {
		o.NoColor = true // Turn off color
	}

	exitCode, err := cli.xfg(o)
	if err != nil {
		return exitErr, fmt.Sprintf("on xfg() : %s", err)
	}

	return exitCode, ""
}

func (cli *runner) xfg(o *options) (int, error) {
	x := newX(cli, o)

	if err := x.search(); err != nil {
		return exitErr, fmt.Errorf("search() : %w", err)
	}

	pagerCloser, err := cli.pager(o.NoPager, x.result.lc)
	if err != nil {
		return exitErr, fmt.Errorf("wrong pgaer : %w", err)
	}
	if pagerCloser != nil {
		defer pagerCloser()
	}

	if err := cli.showResult(x); err != nil {
		return exitErr, fmt.Errorf("showResult() : %w", err)
	}

	return cli.exitCode, nil
}
