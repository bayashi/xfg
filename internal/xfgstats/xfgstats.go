package xfgstats

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"github.com/bayashi/xfg/internal/xfgutil"
)

type lap struct {
	label string
	t     time.Duration
}

type count struct {
	paths        int
	matched      int
	grep         int
	matchedPaths int
	matchedLC    int
}

type Stats struct {
	procs int
	start time.Time
	lap   []lap
	count count
}

func New(procs int) *Stats {
	return &Stats{
		procs: procs,
		start: time.Now(),
	}
}

func (s *Stats) Mark(label string) {
	s.lap = append(s.lap, lap{
		label: label,
		t:     time.Since(s.start),
	})
	s.start = time.Now()
}

func (s *Stats) Show(out io.Writer) {
	result := "[Lap]\n"
	for _, l := range s.lap {
		result = result + fmt.Sprintf(" %s: %s\n", l.label, l.t.String())
	}
	result = result + fmt.Sprintf("[Env]\n procs: %d\n", s.procs)
	result = result + fmt.Sprintf("[Walk]\n paths: %d\n matched: %d\n grep: %d\n", s.count.paths, s.count.matched, s.count.grep)
	result = result + fmt.Sprintf("[Result]\n paths: %d\n lc: %d\n", s.count.matchedPaths, s.count.matchedLC)

	xfgutil.Output(bufio.NewWriter(out), result)
}

func (s *Stats) IncrPaths() {
	s.count.paths++
}

func (s *Stats) IncrMatched() {
	s.count.matched++
}

func (s *Stats) IncrGrep() {
	s.count.grep++
}

func (s *Stats) SetMatchedPaths(count int) {
	s.count.matchedPaths = count
}

func (s *Stats) SetMatchedLC(count int) {
	s.count.matchedLC = count
}
