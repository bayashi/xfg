package main

import (
	"os"
	"path/filepath"
	"testing"

	a "github.com/bayashi/actually"
)

func TestDefaultOptions(t *testing.T) {
	d := defaultOptions()
	a.Got(d).Expect(&options{}).SameType(t)
	a.Got(d.SearchStart).Expect([]string{"."}).Same(t)
	a.Got(d.Indent).Expect(" ").Same(t)
	a.Got(d.GroupSeparator).Expect("--").Same(t)
	a.Got(d.ColorPathBase).Expect("yellow").Same(t)
	a.Got(d.ColorPath).Expect("cyan").Same(t)
	a.Got(d.ColorContent).Expect("red").Same(t)
}

func TestReadRC(t *testing.T) {
	rcFilePath := filepath.Join(t.TempDir(), "test.toml")
	f, _ := os.Create(rcFilePath)
	_, err := f.WriteString("abs = true")
	f.Close()

	a.Got(err).NoError(t)
	t.Setenv(XFG_RC_ENV_KEY, rcFilePath)

	fakeHomeDir := "fake" // not used so far

	o, err := readRC(fakeHomeDir)
	a.Got(err).NoError(t)
	a.Got(o.Abs).True(t)
}

func TestValidateStartPath_Err(t *testing.T) {
	t.Parallel()
	err := validateStartPath([]string{noMatchKeyword})
	a.Got(err).NotNil(t)
	// Linux or Mac: "stat PATH no such file or directory"
	// Windows       "CreateFile PATH The system cannot find the file specified."
	a.Got(err.Error()).Expect(`^(stat|CreateFile) `).Match(t)

	tempDir := t.TempDir()
	tempFilePath := filepath.Join(tempDir, "foo")
	f, _ := os.Create(tempFilePath)
	f.WriteString("123")
	f.Close()

	err = validateStartPath([]string{tempFilePath})
	a.Got(err).NotNil(t)
	a.Got(err.Error()).Expect("path `[^`]+` should point to a directory").Match(t)
}
