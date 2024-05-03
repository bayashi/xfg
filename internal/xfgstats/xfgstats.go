package xfgstats

import (
	"bufio"
	"fmt"
	"io"
	"sync"
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
	pickedLC       int
	outputLC       int
	scannedLC      int
}

type Stats struct {
	mu    sync.RWMutex
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

func (s *Stats) Lock() {
	s.mu.Lock()
}

func (s *Stats) Unlock() {
	s.mu.Unlock()
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
	result = result + fmt.Sprintf("[Walk]\n paths: %d\n contents: %d\n", s.count.walkedPaths, s.count.walkedContents)
	result = result + fmt.Sprintf("[Scanned]\n files: %d\n lines: %d\n", s.count.scannedFile, s.count.scannedLC)
	result = result + fmt.Sprintf("[Result]\n picked paths: %d\n picked lc:%d\n output lc: %d\n", s.count.pickedPaths, s.count.pickedLC, s.count.outputLC)

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

func (s *Stats) AddOutputLC(count int) {
	s.count.outputLC = s.count.outputLC + count
}

func (s *Stats) IncrScannedLC(count int) {
	s.count.scannedLC = s.count.scannedLC + count
}

func (s *Stats) AddPickedLC(i int) {
	s.count.pickedLC = s.count.pickedLC + i
}
