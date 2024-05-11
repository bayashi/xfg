package xfgignore

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/monochromegane/go-gitignore"
)

const (
	GIT                 = "git"
	GITIGNORE_FILE_NAME = ".gitignore"
	XFGIGNORE_FILE_NAME = ".xfgignore"
)

type Matchers []gitignore.IgnoreMatcher

// https://git-scm.com/docs/gitignore

// e.g. $XDG_CONFIG_HOME/git/ignore
func globalXDGGitignorePath() string {
	return filepath.Join(xdg.ConfigHome, GIT, "ignore")
}

// e.g. $HOME/.config/git/ignore
func globalHomeGitignorePath(homeDir string) string {
	return filepath.Join(homeDir, ".config", GIT, "ignore")
}

// e.g. $HOME/.gitignore_global
func globalObsoleteHomeDirGitignorePath(homeDir string) string {
	return filepath.Join(homeDir, GITIGNORE_FILE_NAME+"_global")
}

// e.g. $HOME/.gitignore
func globalHomeDirGitignorePath(homeDir string) string {
	return filepath.Join(homeDir, GITIGNORE_FILE_NAME)
}

// e.g. $REPOSITORY_ROOT/.gitignore
func repoRootDirGitignorePath(rootDir string) string {
	gitCmd, err := exec.LookPath(GIT)
	if err != nil {
		return "" // trap error, probably command not found
	}
	repoRoot, err := exec.Command(gitCmd, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "" // trap error
	}
	absRepoRoot, err := filepath.Abs(string(repoRoot))
	if err != nil {
		return "" // trap error
	}

	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return "" // trap error
	}

	if absRootDir == absRepoRoot {
		return "" // no need to read here. read this later. On each dir.
	}

	return filepath.Join(absRepoRoot, GITIGNORE_FILE_NAME)
}

// e.g. $REPOSITORY_ROOT/info/exclude
func userGitignorePath() string {
	gitCmd, err := exec.LookPath(GIT)
	if err != nil {
		return "" // trap error, probably command not found
	}
	repoRoot, err := exec.Command(gitCmd, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "" // trap error
	}

	return filepath.Join(string(repoRoot), "info", "exclude")
}

// User specifies path
func userConfigGitignorePath() string {
	gitCmd, err := exec.LookPath(GIT)
	if err != nil {
		return "" // trap error, probably command not found
	}
	ret, err := exec.Command(gitCmd, "config", "--get", "core.excludesfile").Output()
	if err != nil {
		return "" // trap error
	}

	return strings.TrimSpace(string(ret))
}

// e.g. $XDG_CONFIG_HOME/xfg/.xfgignore
func userXDGXFGignorePath() string {
	return filepath.Join(xdg.ConfigHome, "xfg", XFGIGNORE_FILE_NAME)
}

// e.g. $HOME/.xfgignore
func userHomeDirXFGignorePath(homeDir string) string {
	return filepath.Join(homeDir, XFGIGNORE_FILE_NAME)
}

func SetUpGlobalGitIgnores(rootDirPath string, homeDir string) Matchers {
	var ms Matchers
	if matcher, err := gitignore.NewGitIgnore(globalXDGGitignorePath(), rootDirPath); err == nil {
		ms = append(ms, matcher)
	} else if matcher, err := gitignore.NewGitIgnore(globalHomeGitignorePath(homeDir), rootDirPath); err == nil {
		ms = append(ms, matcher)
	}

	for _, gitignorePath := range []string{
		globalObsoleteHomeDirGitignorePath(homeDir),
		globalHomeDirGitignorePath(homeDir),
		repoRootDirGitignorePath(rootDirPath),
		userGitignorePath(),
		userConfigGitignorePath(),
	} {
		if gitignorePath == "" {
			continue
		}
		if matcher, err := gitignore.NewGitIgnore(gitignorePath, rootDirPath); err == nil {
			ms = append(ms, matcher)
		}
	}

	return ms
}

func SetupGlobalXFGIgnore(rootDirPath string, homeDir string, xfgignore string) Matchers {
	var ms Matchers
	for _, xfgignorePath := range []string{
		userXDGXFGignorePath(),
		userHomeDirXFGignorePath(homeDir),
		xfgignore,
	} {
		if matcher, err := gitignore.NewGitIgnore(xfgignorePath, rootDirPath); err == nil {
			ms = append(ms, matcher)
		}
	}

	return ms
}
