package bootstrap

import (
	"os"
	"os/exec"
)

// Exec runs args as a child process wired to the entrypoint standard
// streams; Windows has no execve equivalent.
func Exec(args []string, env Environment) error {
	command := exec.Command(args[0], args[1:]...)
	command.Env = env.List()
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}
