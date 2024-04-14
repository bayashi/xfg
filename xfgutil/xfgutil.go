package xfgutil

import (
	"bufio"
	"fmt"
	"os"
	"runtime"

	"github.com/mattn/go-isatty"
	"golang.org/x/term"
)

func Procs() int {
	cpu := runtime.NumCPU()
	if cpu == 1 {
		cpu = 2
	}

	runtime.GOMAXPROCS(cpu)

	return cpu
}

func IsTTY() bool {
	fd := os.Stdout.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

func HomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return homeDir, nil
}

func GetTermWindowRows(fd int) (int, error) {
	_, rows, err := term.GetSize(fd)
	if err != nil {
		return 0, err
	}

	return rows, nil
}

func Output(writer *bufio.Writer, out string) error {
	if _, err := fmt.Fprint(writer, out); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}
