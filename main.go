package main

import (
	"fmt"
	"io"
	"os"
)

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
	os.Exit(exitOK)
}

func (cli *runner) run() {
	err := cli.xfg(cli.parseArgs())
	if err != nil {
		cli.putErr(fmt.Sprintf("Err: %s", err))
		os.Exit(exitErr)
	}
}

func (cli *runner) xfg(o *options) error {
	x := newX(o)

	if err := x.search(); err != nil {
		return fmt.Errorf("error during Search %w", err)
	}

	if err := x.showResult(cli.out); err != nil {
		return fmt.Errorf("error during Show %w", err)
	}

	return nil
}
