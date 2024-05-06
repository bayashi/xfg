package xfgignore

import (
	"testing"

	a "github.com/bayashi/actually"
)

func TestSetUpGlobalGitIgnores(t *testing.T) {
	a.Got(SetUpGlobalGitIgnores("", "")).Expect(Matchers{}).SameType(t)
}

func TestSetupGlobalXFGIgnore(t *testing.T) {
	a.Got(SetupGlobalXFGIgnore("", "", "")).Expect(Matchers{}).SameType(t)
}
