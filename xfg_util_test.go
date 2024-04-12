package main

import (
	"os"
	"path/filepath"
	"testing"

	a "github.com/bayashi/actually"
)

func TestIsTTY(t *testing.T) {
	// testing envs are non-TTY, comomnly. Local and Github Actions are non-TTY.
	a.Got(isTTY()).False(t)
}

func TestHomeDir(t *testing.T) {
	homeDir, err := homeDir()
	a.Got(err).NoError(t)
	a.Got(homeDir).Expect("").NotSame(t)
}

func TestReadRC(t *testing.T) {
	rcFilePath := filepath.Join(t.TempDir(), "test.toml")
	f, _ := os.Create(rcFilePath)
	_, err := f.WriteString("relax = true")
	f.Close()

	a.Got(err).NoError(t)
	t.Setenv(XFG_RC_ENV_KEY, rcFilePath)

	fakeHomeDir := "fake" // not used so far

	o, err := readRC(fakeHomeDir)
	a.Got(err).NoError(t)
	a.Got(o.Relax).True(t)
}

func TestPrepareGitIgnore(t *testing.T) {
	tempDir := t.TempDir()
	gitignoreFilePath := filepath.Join(tempDir, ".gitignore")
	f, _ := os.Create(gitignoreFilePath)
	f.WriteString("ignorez")
	f.Close()

	gitignore := prepareGitIgnore("", tempDir)

	a.Got(gitignore.MatchesPath("ignorez")).True(t)
}

func TestPrepareXfgIgnore(t *testing.T) {
	tempDir := t.TempDir()
	xfgignoreFilePath := filepath.Join(tempDir, ".xfgignore")
	f, err := os.Create(xfgignoreFilePath)
	a.Got(err).NoError(t)
	f.WriteString("ignorex")
	f.Close()

	xfgignore := prepareXfgIgnore("", xfgignoreFilePath)
	a.Got(xfgignore).NotNil(t)

	a.Got(xfgignore.MatchesPath("ignorex")).True(t)
}

func TestValidateStartPath_Err(t *testing.T) {
	err := validateStartPath(noMatchKeyword)
	a.Got(err).NotNil(t)
	a.Got(err.Error()).Expect(`^wrong path `).Match(t)

	tempDir := t.TempDir()
	tempFilePath := filepath.Join(tempDir, "foo")
	f, _ := os.Create(tempFilePath)
	f.WriteString("123")
	f.Close()

	err = validateStartPath(tempFilePath)
	a.Got(err).NotNil(t)
	a.Got(err.Error()).Expect("path `[^`]+` should point to a directory").Match(t)
}

func TestProcs(t *testing.T) {
	a.Got(procs() > 0).True(t)
}
