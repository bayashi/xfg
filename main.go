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
	out   io.Writer
	err   io.Writer
	isTTY bool
}

func main() {
	stdout := os.Stdout
	fd := stdout.Fd()
	cli := &runner{
		out:   stdout,
		err:   os.Stderr,
		isTTY: isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd),
	}
	cli.run()
	funcExit(exitOK)
}

func (cli *runner) run() {
	o := cli.parseArgs()
	if !cli.isTTY {
		o.noColor = true // Turn off color
	}

	err := cli.xfg(o)
	if err != nil {
		cli.putErr(fmt.Sprintf("Err: %s", err))
		funcExit(exitErr)
	}
}

func (cli *runner) xfg(o *options) error {
	o.prepareContextLines()
	x := newX(o)

	if err := x.search(); err != nil {
		return fmt.Errorf("error during Search %w", err)
	}

	closer, err := cli.pager(o.noPager, x.resultLines)
	if err != nil {
		return fmt.Errorf("pgaer is wrong: %w", err)
	}
	if closer != nil {
		defer closer()
	}

	if err := cli.showResult(x); err != nil {
		return fmt.Errorf("error during Show %w", err)
	}

	return nil
}
