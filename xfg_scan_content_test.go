package main

import (
	"bufio"
	"strings"
	"testing"

	a "github.com/bayashi/actually"
	"github.com/bayashi/xfg/internal/xfgstats"
)

func newScanTestXfg(o *options) *xfg {
	o.prepareContextLines(true)
	return &xfg{
		cli: &runner{
			stats: xfgstats.New(1),
		},
		options: o,
	}
}

func scanTestContent(t *testing.T, x *xfg, content string) []line {
	t.Helper()
	got, err := x.scanContent(bufio.NewScanner(strings.NewReader(content)), "test")
	a.Got(err).NoError(t)
	return got
}

func linesToPairs(lines []line) [][2]any {
	out := make([][2]any, 0, len(lines))
	for _, l := range lines {
		out = append(out, [2]any{l.lc, l.content})
	}
	return out
}

func TestScanContent_BeforeContextPartialWindow(t *testing.T) {
	t.Parallel()
	x := newScanTestXfg(&options{
		SearchGrep:         []string{"HIT"},
		BeforeContextLines: 3,
	})
	got := scanTestContent(t, x, "a\nHIT\nc\n")
	a.Got(linesToPairs(got)).Expect([][2]any{
		{int32(1), "a"},
		{int32(2), "HIT"},
	}).Same(t)
	a.Got(got[1].matched).True(t)
}

func TestScanContent_BeforeContextKeepsOnlyRecentLines(t *testing.T) {
	t.Parallel()
	x := newScanTestXfg(&options{
		SearchGrep:         []string{"HIT"},
		BeforeContextLines: 2,
	})
	content := strings.Join([]string{
		"old1", "old2", "old3", "old4", "old5",
		"keep1", "keep2", "HIT", "after",
	}, "\n") + "\n"
	got := scanTestContent(t, x, content)
	a.Got(linesToPairs(got)).Expect([][2]any{
		{int32(6), "keep1"},
		{int32(7), "keep2"},
		{int32(8), "HIT"},
	}).Same(t)
}

func TestScanContent_BeforeContextClearedBetweenMatches(t *testing.T) {
	t.Parallel()
	x := newScanTestXfg(&options{
		SearchGrep:         []string{"HIT"},
		BeforeContextLines: 1,
	})
	content := strings.Join([]string{
		"b1", "HIT",
		"mid1", "mid2",
		"b2", "HIT",
	}, "\n") + "\n"
	got := scanTestContent(t, x, content)
	a.Got(linesToPairs(got)).Expect([][2]any{
		{int32(1), "b1"},
		{int32(2), "HIT"},
		{int32(5), "b2"},
		{int32(6), "HIT"},
	}).Same(t)
	a.Got(got[1].matched).True(t)
	a.Got(got[3].matched).True(t)
	a.Got(got[0].matched).False(t)
	a.Got(got[2].matched).False(t)
}

func TestScanContent_BeforeAndAfterContext(t *testing.T) {
	t.Parallel()
	x := newScanTestXfg(&options{
		SearchGrep:         []string{"HIT"},
		BeforeContextLines: 1,
		AfterContextLines:  1,
	})
	content := strings.Join([]string{
		"b", "HIT", "a",
		"gap1", "gap2",
		"b2", "HIT", "a2",
	}, "\n") + "\n"
	got := scanTestContent(t, x, content)
	a.Got(linesToPairs(got)).Expect([][2]any{
		{int32(1), "b"},
		{int32(2), "HIT"},
		{int32(3), "a"},
		{int32(6), "b2"},
		{int32(7), "HIT"},
		{int32(8), "a2"},
	}).Same(t)
}

func TestScanContent_MaxColumns(t *testing.T) {
	t.Parallel()
	x := newScanTestXfg(&options{
		SearchGrep: []string{"HIT"},
		MaxColumns: 3,
	})
	got := scanTestContent(t, x, "HIT-EXTRA\n")
	a.Got(len(got)).Expect(1).Same(t)
	a.Got(got[0].content).Expect("HIT").Same(t)
}
