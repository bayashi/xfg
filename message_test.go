package main

import "testing"

func TestGetMessage(t *testing.T) {
	getMessage("help_Stats")

	t.Setenv("LANG", "ja_JP.UTF-8")
	getMessage("help_Stats")
}
