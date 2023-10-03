package xdeb

import (
	"os"
	"os/exec"
	"strings"
)

func ExecuteCommand(workdir string, args ...string) error {
	LogMessage("Executing command: %s ...", strings.Join(args, " "))

	command := exec.Command(args[0], args[1:]...)
	command.Dir = workdir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	command.Env = append(command.Environ(), "XDEB_PKGROOT=" + workdir)

	return command.Run()
}
