package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Entrypoint is the testable core of a container entrypoint. Main supplies the
// process environment and command-line arguments.
type Entrypoint func(Environment, []string) error

// Main runs an entrypoint with the current process environment and arguments,
// then terminates the process if it fails.
func Main(entrypoint Entrypoint) {
	env := NewEnvironment(os.Environ())
	ExitOnError(entrypoint(env, os.Args[1:]))
}

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
	if err == nil {
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

func resolveCommand(args []string, binary string) []string {
	if serviceArgs, startsService := ServiceArgs(args, binary); startsService {
		return append([]string{binary}, serviceArgs...)
	}

	return args
}

// Execute hands control over to args with the given environment.
func Execute(args []string, env Environment) error {
	if len(args) == 0 {
		return errors.New("execute: empty command")
	}

	if err := executeProcess(args, env); err != nil {
		return fmt.Errorf("execute %s: %w", args[0], err)
	}

	return nil
}

// Service returns an entrypoint which prepares service when the container
// is about to start its default binary. Custom user commands are executed
// untouched.
func Service(binary string, prepare func(Environment) error) Entrypoint {
	return func(env Environment, args []string) error {
		command := resolveCommand(args, binary)

		if command[0] == binary {
			if err := prepare(env); err != nil {
				return err
			}
		}

		return Execute(command, env)
	}
}

const (
	initDBCommand               = "init_db_only"
	initHistoryProvidersCommand = "init_db_history_providers"
	initTelemetryCommand        = "init_db_telemetry"
)

// DBService is Service for images with a database. The extra initDB hook
// implements the initDBCommand command.
func DBService(binary string, prepare, initDB func(Environment) error) Entrypoint {
	return func(env Environment, args []string) error {
		command := resolveCommand(args, binary)

		switch command[0] {
		case binary:
			if err := prepare(env); err != nil {
				return err
			}
		case initDBCommand:
			return initDB(env)
		}

		return Execute(command, env)
	}
}

// StorageInitEntrypoint handles independent database initialization commands without starting the service.
// A nil handler marks a command as unsupported by the image.
func StorageInitEntrypoint(next Entrypoint, history, telemetry func(Environment) (int, error)) Entrypoint {
	return func(env Environment, args []string) error {
		if len(args) == 0 {
			return next(env, args)
		}

		var prepare func(Environment) (int, error)

		switch args[0] {
		case initHistoryProvidersCommand:
			prepare = history
		case initTelemetryCommand:
			prepare = telemetry
		default:
			return next(env, args)
		}

		if prepare == nil {
			return fmt.Errorf("%s is not supported by this image", args[0])
		}

		if len(args) != 1 {
			return fmt.Errorf("%s takes no arguments", args[0])
		}

		count, err := prepare(env)
		if err != nil {
			return err
		}
		if count == 0 {
			if args[0] == initHistoryProvidersCommand {
				return errors.New("no history schemas initialized: set ZBX_HISTORYPROVIDER_<N>_SCHEMA for the providers to initialize")
			}
			return errors.New("no telemetry schemas initialized: ZBX_TELEMETRYPROVIDER_0_SCHEMA is not set")
		}

		return nil
	}
}
