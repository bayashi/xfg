package main

import (
	"os"
	"path/filepath"
	"testing"

	a "github.com/bayashi/actually"
)

func TestReadRC(t *testing.T) {
	rcFilePath := filepath.Join(t.TempDir(), "test.toml")
	f, _ := os.Create(rcFilePath)
	defer f.Close()
	_, err := f.WriteString("relax = true")
	a.Got(err).NoError(t)
	t.Setenv(XFG_RC_ENV_KEY, rcFilePath)

	o, err := readRC()
	a.Got(err).NoError(t)
	a.Got(o.Relax).True(t)
}
