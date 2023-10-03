package xdeb

import (
	"fmt"
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

	if args[0] == "xdeb" {
		command.Env = append(command.Env, fmt.Sprintf("XDEB_PKGROOT=%s", workdir))
	}

	return command.Run()
}
