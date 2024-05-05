package main

import (
	"fmt"
	"io"
	"os"

	"github.com/bayashi/xfg/internal/xfgpager"
	"github.com/bayashi/xfg/internal/xfgstats"
	"github.com/bayashi/xfg/internal/xfgutil"
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
	procs    int
	stats    *xfgstats.Stats
}

func main() {
	procs := xfgutil.Procs()
	cli := &runner{
		out:   os.Stdout,
		err:   os.Stderr,
		isTTY: xfgutil.IsTTY(),
		procs: procs,
		stats: xfgstats.New(procs),
	}
	exitCode, message := cli.run()

	if exitCode != exitOK {
		cli.putErr(fmt.Sprintf("Err: %s", message))
	}
	funcExit(exitCode)
}

func (cli *runner) run() (int, string) {
	if o, err := cli.preXfg(); err != nil {
		return exitErr, fmt.Sprintf("on preXfg() : %s", err)
	} else {
		exitCode, err := cli.xfg(o)
		if err != nil {
			return exitErr, fmt.Sprintf("on xfg() : %s", err)
		}

		return exitCode, ""
	}
}

func (cli *runner) preXfg() (*options, error) {
	homeDir, err := xfgutil.HomeDir()
	if err != nil {
		return nil, fmt.Errorf("on detecting home directory : %s", err)
	}
	cli.homeDir = homeDir

	cli.stats.Mark("homeDir")

	defaultOpt, err := readRC(cli.homeDir)
	if err != nil {
		return nil, fmt.Errorf("on reading home directory : %s", err)
	}

	cli.stats.Mark("readRC")

	o := cli.parseArgs(defaultOpt)
	if !cli.isTTY {
		o.NoColor = true // Turn off color
	}

	if o.Stats {
		cli.stats.Mark("parseArgs")
	}

	if err := o.validateOptions(); err != nil {
		return nil, err
	}

	return o, nil
}

func (cli *runner) xfg(o *options) (int, error) {
	x := newX(cli, o)

	if err := x.process(); err != nil {
		return exitErr, fmt.Errorf("process() : %w", err)
	}

	if x.options.Stats {
		cli.stats.Mark("process")
	}

	if !x.options.NoPager && cli.isTTY {
		out, pagerCloser, err := xfgpager.Pager(cli.out, cli.err, x.result.outputLC)
		if err != nil {
			return exitErr, fmt.Errorf("wrong pgaer : %w", err)
		}
		if pagerCloser != nil {
			defer pagerCloser()
		}
		if out != nil {
			cli.out = out
		}
	}

	if x.options.Stats {
		cli.stats.Mark("pager")
	}

	if err := cli.showResult(x); err != nil {
		return exitErr, fmt.Errorf("showResult() : %w", err)
	}

	if x.options.Stats {
		cli.stats.Mark("showResult")
		cli.stats.SetPickedPaths(len(x.result.paths))

		cli.stats.Show(cli.out)
	}

	return cli.exitCode, nil
}
