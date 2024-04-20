package xfglangxt

import (
	"testing"

	a "github.com/bayashi/actually"
)

func TestIssupported(t *testing.T) {
	a.Got(IsSupported("perl")).True(t)
	a.Got(IsSupported("Perl")).True(t)
	a.Got(IsSupported("english")).False(t)
}

func TestGet(t *testing.T) {
	a.Got(Get("perl")).Expect([]string{".pl", ".pm", ".t", ".pod", ".PL"}).Same(t)
	a.Got(Get("Perl")).Expect([]string{".pl", ".pm", ".t", ".pod", ".PL"}).Same(t)
}

func TestIsLangFile(t *testing.T) {
	a.Got(IsLangFile("perl", "foo.pl")).True(t)
	a.Got(IsLangFile("php", "foo.pl")).False(t)
}
