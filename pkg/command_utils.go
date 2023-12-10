package xdeb

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func findArgumentIndex(command string, args ...string) int {
	for i := 0; i < len(args); i++ {
		if strings.HasSuffix(args[i], command) {
			return i
		}
	}

	return -1
}

func ExecuteCommand(workdir string, args ...string) error {
	LogMessage("Executing command: %s ...", strings.Join(args, " "))

	command := exec.Command(args[0], args[1:]...)
	command.Dir = workdir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	commandIndex := findArgumentIndex("xdeb", args...)

	if commandIndex > -1 {
		commandStringSlice := strings.Split(args[commandIndex], "/")

		// Add xdeb-specific environment to xdeb command
		if commandStringSlice[len(commandStringSlice)-1] == "xdeb" {
			for _, environmentVariable := range os.Environ() {
				if strings.HasPrefix(environmentVariable, "XDEB_") {
					command.Env = append(command.Env, environmentVariable)
				}
			}

			command.Env = append(command.Env, fmt.Sprintf("XDEB_PKGROOT=%s", workdir))
		}
	}

	command.Env = append(command.Env, fmt.Sprintf("PATH=%s", os.Getenv("PATH")))
	return command.Run()
}
