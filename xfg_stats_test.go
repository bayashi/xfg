package main

import (
	"bytes"
	"testing"

	a "github.com/bayashi/actually"
)

func TestStats(t *testing.T) {
	stats := newStats(1)
	stats.mark("step1")
	stats.mark("step2")
	var o bytes.Buffer
	stats.show(&o)
	a.Got(o.String()).Expect(`\nstep1:\s+\w+\nstep2:\s+\w+\nprocs:\s+\d+\npaths:\s+0\nmatched:\s+0\ngrep:\s+0\n`).Match(t)
}
