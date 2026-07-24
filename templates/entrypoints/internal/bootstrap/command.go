package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// exitCode maps err to a process exit status, preserving the status of a
// finished child process.
func exitCode(err error) int {
	if err == nil {
		return 0
	}

	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}

	return 1
}

// ExitOnError logs err and terminates the process with a matching exit
// status. A nil error is ignored.
func ExitOnError(err error) {
	if err == nil || errors.Is(err, context.Canceled) {
		return
	}

	LogError("**** %v", err)
	os.Exit(exitCode(err))
}

// ServiceArgs returns arguments intended for service and whether the service
// should be started. Empty arguments, the service command itself and options
// beginning with a dash select the service; anything else is a custom command.
func ServiceArgs(args []string, service string) ([]string, bool) {
	if len(args) == 0 {
		return nil, true
	}
	if args[0] == service {
		return args[1:], true
	}
	if strings.HasPrefix(args[0], "-") {
		return args, true
	}

	return nil, false
}

// resolveCommand decides what the container should execute.
func resolveCommand(args []string, binary string) []string {
	if serviceArgs, startsService := ServiceArgs(args, binary); startsService {
		return append([]string{binary}, serviceArgs...)
	}

	return args
}

// Execute hands control over to args with the given environment.
func Execute(args []string, env Environment) error {
	if err := executeProcess(args, env); err != nil {
		return fmt.Errorf("execute %s: %w", args[0], err)
	}

	return nil
}

// RunService implements the common entrypoint flow: when the container is
// about to start the service binary (the image default), prepare is called
// first; custom user commands are executed untouched.
func RunService(binary string, prepare func(Environment) error) error {
	env := NewEnvironment(os.Environ())
	args := resolveCommand(os.Args[1:], binary)

	if args[0] == binary {
		if err := prepare(env); err != nil {
			return err
		}
	}

	return Execute(args, env)
}

// initDBCommand provisions the database and exits without starting
// the service.
const initDBCommand = "init_db_only"

// RunDBService is RunService for images with a database: the extra
// initDB hook implements the initDBCommand command.
func RunDBService(binary string, prepare, initDB func(Environment) error) error {
	env := NewEnvironment(os.Environ())
	args := resolveCommand(os.Args[1:], binary)

	switch args[0] {
	case binary:
		if err := prepare(env); err != nil {
			return err
		}
	case initDBCommand:
		return initDB(env)
	}

	return Execute(args, env)
}
