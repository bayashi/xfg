package xfgpager

import (
	"bytes"
	"io"
	"testing"

	a "github.com/bayashi/actually"
)

func TestPager(t *testing.T) {
	var o bytes.Buffer
	var e bytes.Buffer

	GetTermWindowRows = func() (int, error) { return 10, nil }

	w, closer, err := Pager(&o, &e, 10)
	a.Got(err).NoError(t)
	a.Got(closer).NotNil(t)
	a.Got(w).Expect(&io.PipeWriter{}).SameType(t)
}

func TestPagerEnv(t *testing.T) {
	var o bytes.Buffer
	var e bytes.Buffer

	GetTermWindowRows = func() (int, error) { return 10, nil }

	// No Pager
	t.Setenv(XFG_PAGER_ENV_KEY, "NOPAGER")
	w, closer, err := Pager(&o, &e, 10)
	a.Got(err).NoError(t)
	a.Got(closer).Nil(t)
	a.Got(w).Nil(t)

	// With Pager
	t.Setenv(XFG_PAGER_ENV_KEY, "")
	w, closer, err = Pager(&o, &e, 10)
	a.Got(err).NoError(t)
	a.Got(closer).NotNil(t)
	a.Got(w).Expect(&io.PipeWriter{}).SameType(t)
}
