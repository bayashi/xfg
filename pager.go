package main

/*
	The most part of this file `pager.go` was copied from https://github.com/jackdoe/go-pager
	Copyright © 2020, Borislav Nikolov
	All rights reserved.
*/

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/bayashi/xfg/xfgutil"
)

const (
	ENV_KEY_PAGER      = "PAGER"
	ENV_VALUE_NO_PAGER = "NOPAGER"
)

func (cli *runner) pager(noPager bool, result int) (func(), error) {
	if !cli.isTTY || noPager {
		return nil, nil
	}

	rows, err := xfgutil.GetTermWindowRows(int(syscall.Stdout))
	if err != nil {
		return nil, err
	}
	if rows-1 > result {
		return nil, nil // No need pager. Don't think about gourp separators
	}

	p, err := pagerPath("less", "more", "lv")
	if err != nil {
		return nil, err
	}

	if p != "" {
		signal.Ignore(syscall.SIGPIPE)

		c := exec.Command(p)
		r, w := io.Pipe()
		c.Stdin = r
		c.Stdout = cli.out
		c.Stderr = cli.err
		ch := make(chan struct{})
		go func() {
			defer close(ch)
			err := c.Run()
			if err != nil {
				panic(err)
			}
			os.Exit(exitOK)
		}()

		cli.out = w

		return func() {
			w.Close()
			<-ch
		}, nil
	}

	return nil, nil
}

func pagerPath(pagers ...string) (string, error) {
	pager := os.Getenv(ENV_KEY_PAGER)
	if pager != "" {
		if pager == ENV_VALUE_NO_PAGER {
			return "", nil
		}

		exe, err := exec.LookPath(pager)
		if err != nil {
			return "", fmt.Errorf("could not execute `%s` from ENV:%s: %w", pager, ENV_KEY_PAGER, err)
		}
		return exe, nil
	}

	for _, p := range pagers {
		exe, err := exec.LookPath(p)
		if err != nil {
			return "", fmt.Errorf("could not execute PAGER:%s: %w", p, err)
		} else {
			return exe, nil
		}
	}

	return "", nil
}
