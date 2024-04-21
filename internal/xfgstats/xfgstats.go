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
	walkedPaths    int
	walkedContents int
	scannedFile    int
	pickedPaths    int
	totalLC        int
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
	result = result + fmt.Sprintf("[Walk]\n paths: %d\n matched: %d\n grep: %d\n", s.count.walkedPaths, s.count.walkedContents, s.count.scannedFile)
	result = result + fmt.Sprintf("[Result]\n paths: %d\n lc: %d\n", s.count.pickedPaths, s.count.totalLC)

	xfgutil.Output(bufio.NewWriter(out), result)
}

func (s *Stats) IncrWalkedPaths() {
	s.count.walkedPaths++
}

func (s *Stats) IncrWalkedContents() {
	s.count.walkedContents++
}

func (s *Stats) IncrScannedFile() {
	s.count.scannedFile++
}

func (s *Stats) SetPickedPaths(count int) {
	s.count.pickedPaths = count
}

func (s *Stats) SetTotalLC(count int) {
	s.count.totalLC = count
}
