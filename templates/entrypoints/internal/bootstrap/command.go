package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}

	return 1
}

func ExitOnError(err error) {
	if err == nil {
		return
	}

	LogError("**** %v", err)
	os.Exit(ExitCode(err))
}

func Command(args []string, binary string) []string {
	if len(args) == 0 {
		return []string{binary}
	}
	if strings.HasPrefix(args[0], "-") {
		return append([]string{binary}, args...)
	}
	return args
}

func Execute(args []string, env Environment) error {
	if err := Exec(args, env); err != nil {
		return fmt.Errorf("execute %s: %w", args[0], err)
	}

	return nil
}

func RunService(binary string, prepare func(Environment) error) error {
	env := NewEnvironment(os.Environ())
	args := Command(os.Args[1:], binary)

	if args[0] == binary {
		if err := prepare(env); err != nil {
			return err
		}
	}

	return Execute(args, env)
}
