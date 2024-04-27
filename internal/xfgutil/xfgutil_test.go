package xfgutil

import (
	"testing"

	a "github.com/bayashi/actually"
)

func TestProcs(t *testing.T) {
	t.Parallel()
	a.Got(Procs() > 0).True(t)
}

func TestIsTTY(t *testing.T) {
	t.Parallel()
	// testing envs are non-TTY, comomnly. Local and Github Actions are non-TTY.
	a.Got(IsTTY()).False(t)
}

func TestHomeDir(t *testing.T) {
	t.Parallel()
	homeDir, err := HomeDir()
	a.Got(err).NoError(t)
	a.Got(homeDir).Expect("").NotSame(t)
}
