package main

import (
	"fmt"
	"io"
	"os"
)

var funcExit = func(code int) {
	os.Exit(code)
}

type runner struct {
	out io.Writer
	err io.Writer
}

func main() {
	cli := &runner{
		out: os.Stdout,
		err: os.Stderr,
	}
	cli.run()
	funcExit(exitOK)
}

func (cli *runner) run() {
	err := cli.xfg(cli.parseArgs())
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

	if err := x.showResult(cli.out); err != nil {
		return fmt.Errorf("error during Show %w", err)
	}

	return nil
}
