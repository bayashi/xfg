package xfgstats

import (
	"bytes"
	"testing"

	a "github.com/bayashi/actually"
)

func TestStats(t *testing.T) {
	stats := New(1)
	stats.Mark("step1")
	stats.Mark("step2")
	var o bytes.Buffer
	stats.Show(&o)
	a.Got(o.String()).Expect(`\nstep1:\s+.+\nstep2:\s+.+\nprocs:\s+1\npaths:\s+0\nmatched:\s+0\ngrep:\s+0\n`).Match(t)
}
