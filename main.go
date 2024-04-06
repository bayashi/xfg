package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

var funcExit = func(code int) {
	os.Exit(code)
}

type runner struct {
	out      io.Writer
	err      io.Writer
	isTTY    bool
	exitCode int
}

func main() {
	stdout := os.Stdout
	fd := stdout.Fd()
	cli := &runner{
		out:   stdout,
		err:   os.Stderr,
		isTTY: isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd),
	}
	exitCode := cli.run()
	funcExit(exitCode)
}

func (cli *runner) run() int {
	defaultOpt, err := readRC()
	if err != nil {
		cli.putErr(fmt.Sprintf("Err: on reading config : %s", err))
		funcExit(exitErr)
	}
	o := cli.parseArgs(defaultOpt)
	if !cli.isTTY {
		o.NoColor = true // Turn off color
	}

	exitCode, err := cli.xfg(o)
	if err != nil {
		cli.putErr(fmt.Sprintf("Err: %s", err))
		funcExit(exitErr)
	}

	return exitCode
}

func (cli *runner) xfg(o *options) (int, error) {
	o.prepareContextLines(cli.isTTY)
	x := newX(o)

	if err := x.search(); err != nil {
		return exitErr, fmt.Errorf("error during Search %w", err)
	}

	closer, err := cli.pager(o.NoPager, x.resultLines)
	if err != nil {
		return exitErr, fmt.Errorf("pgaer is wrong: %w", err)
	}
	if closer != nil {
		defer closer()
	}

	if err := cli.showResult(x); err != nil {
		return exitErr, fmt.Errorf("error during Show %w", err)
	}

	return cli.exitCode, nil
}
