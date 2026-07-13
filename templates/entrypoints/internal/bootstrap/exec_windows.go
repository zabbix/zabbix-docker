package bootstrap

import (
	"os"
	"os/exec"
)

func Exec(args []string, env Environment) error {
	command := exec.Command(args[0], args[1:]...)
	command.Env = env.List()
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}
