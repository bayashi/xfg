package main

import (
	"bufio"
	"fmt"
	"io"
	"time"
)

type lap struct {
	label string
	t     time.Duration
}

type count struct {
	paths   int
	matched int
	grep    int
}

type stats struct {
	procs int
	start time.Time
	lap   []lap
	count count
}

func newStats(procs int) *stats {
	return &stats{
		procs: procs,
		start: time.Now(),
	}
}

func (s *stats) mark(label string) {
	s.lap = append(s.lap, lap{
		label: label,
		t:     time.Since(s.start),
	})
	s.start = time.Now()
}

func (s *stats) show(out io.Writer) {
	result := "\n"
	for _, l := range s.lap {
		result = result + fmt.Sprintf("%s: %s\n", l.label, l.t.String())
	}

	result = result + fmt.Sprintf("procs: %d\n", s.procs)
	result = result + fmt.Sprintf("paths: %d\nmatched: %d\ngrep: %d\n", s.count.paths, s.count.matched, s.count.grep)

	output(bufio.NewWriter(out), result)
}
