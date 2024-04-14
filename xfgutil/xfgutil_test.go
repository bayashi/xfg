package xfgutil

import (
	"testing"

	a "github.com/bayashi/actually"
)

func TestProcs(t *testing.T) {
	a.Got(Procs() > 0).True(t)
}

func TestIsTTY(t *testing.T) {
	// testing envs are non-TTY, comomnly. Local and Github Actions are non-TTY.
	a.Got(IsTTY()).False(t)
}

func TestHomeDir(t *testing.T) {
	homeDir, err := HomeDir()
	a.Got(err).NoError(t)
	a.Got(homeDir).Expect("").NotSame(t)
}
