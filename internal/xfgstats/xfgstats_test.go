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
	a.Got(o.String()).Expect(`\[Lap\]\n`).Match(t)
	a.Got(o.String()).Expect(`step1:\s+.+\n`).Match(t)
	a.Got(o.String()).Expect(`step2:\s+.+\n`).Match(t)
	a.Got(o.String()).Expect(`\[Env\]\n`).Match(t)
	a.Got(o.String()).Expect(`procs:\s+\d+\n`).Match(t)
	a.Got(o.String()).Expect(`\[Walk\]\n`).Match(t)
	a.Got(o.String()).Expect(`paths:\s+\d+\n`).Match(t)
	a.Got(o.String()).Expect(`matched:\s+\d+\n`).Match(t)
	a.Got(o.String()).Expect(`grep:\s+\d+\n`).Match(t)
	a.Got(o.String()).Expect(`\[Result\]\n`).Match(t)
	a.Got(o.String()).Expect(`paths:\s+\d+\n`).Match(t)
	a.Got(o.String()).Expect(`lc:\s+\d\n`).Match(t)
}
