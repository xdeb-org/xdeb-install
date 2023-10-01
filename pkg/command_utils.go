package xdeb

import (
	"os"
	"os/exec"
)

func ExecuteCommand(workdir string, args ...string) error {
	command := exec.Command(args[0], args[1:]...)
	command.Dir = workdir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	return command.Run()
}
