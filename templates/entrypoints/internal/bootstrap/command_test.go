package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestResolveCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "default", want: []string{"component"}},
		{name: "component options", args: []string{"--version"}, want: []string{"component", "--version"}},
		{name: "custom command", args: []string{"shell", "argument"}, want: []string{"shell", "argument"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := resolveCommand(test.args, "component"); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("resolveCommand() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestServiceArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		want       []string
		wantStarts bool
	}{
		{name: "empty", wantStarts: true},
		{name: "service command", args: []string{"component"}, want: []string{}, wantStarts: true},
		{
			name:       "service options",
			args:       []string{"component", "--debug"},
			want:       []string{"--debug"},
			wantStarts: true,
		},
		{name: "options", args: []string{"--version"}, want: []string{"--version"}, wantStarts: true},
		{name: "custom command", args: []string{"shell", "argument"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, starts := ServiceArgs(test.args, "component")
			if starts != test.wantStarts || !reflect.DeepEqual(got, test.want) {
				t.Fatalf("ServiceArgs(%q) = %q, %t; want %q, %t",
					test.args, got, starts, test.want, test.wantStarts)
			}
		})
	}
}

func TestExitCode(t *testing.T) {
	if code := exitCode(nil); code != 0 {
		t.Fatalf("exitCode(nil) = %d, want 0", code)
	}
	if code := exitCode(errors.New("failure")); code != 1 {
		t.Fatalf("exitCode(regular error) = %d, want 1", code)
	}

	command := exec.Command(os.Args[0], "-test.run=TestExitCodeHelperProcess")
	command.Env = append(os.Environ(), "ENTRYPOINT_EXIT_CODE_HELPER=1")
	err := command.Run()
	if code := exitCode(err); code != 23 {
		t.Fatalf("exitCode(child error) = %d, want 23: %v", code, err)
	}
}

func TestExitCodeHelperProcess(t *testing.T) {
	if os.Getenv("ENTRYPOINT_EXIT_CODE_HELPER") != "1" {
		return
	}

	os.Exit(23)
}

func TestInitDBExitStatus(t *testing.T) {
	for _, test := range []struct {
		name string
		want int
	}{
		{name: "success", want: 0},
		{name: "failure", want: 1},
		{name: "canceled", want: 1},
		{name: "wrapped canceled", want: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			command := exec.Command(os.Args[0], "-test.run=^TestInitDBExitStatusHelperProcess$")
			command.Env = append(os.Environ(), "ENTRYPOINT_INIT_DB_EXIT_HELPER="+test.name)
			output, err := command.CombinedOutput()
			if command.ProcessState == nil {
				t.Fatalf("start init-only helper: %v", err)
			}
			if got := command.ProcessState.ExitCode(); got != test.want {
				t.Fatalf("init-only exit code = %d, want %d: %v\n%s", got, test.want, err, output)
			}
		})
	}
}

func TestInitDBExitStatusHelperProcess(t *testing.T) {
	mode := os.Getenv("ENTRYPOINT_INIT_DB_EXIT_HELPER")
	if mode == "" {
		return
	}

	var err error
	switch mode {
	case "success":
	case "failure":
		err = errors.New("initialization failed")
	case "canceled":
		err = context.Canceled
	case "wrapped canceled":
		err = fmt.Errorf("wait for database: %w", context.Canceled)
	default:
		t.Fatalf("unknown helper mode %q", mode)
	}

	os.Args = []string{os.Args[0], initDBCommand}
	Main(DBService("component",
		func(Environment) error { return errors.New("unexpected regular preparation") },
		func(Environment) error { return err },
	))
	os.Exit(0)
}

func TestExecuteRejectsEmptyCommand(t *testing.T) {
	err := Execute(nil, Environment{})
	if err == nil || err.Error() != "execute: empty command" {
		t.Fatalf("Execute(nil) error = %v, want empty command error", err)
	}
}

func TestServiceReturnsPreparationError(t *testing.T) {
	want := errors.New("prepare failed")
	env := Environment{"TEST_VALUE": "preserved"}

	err := Service("component", func(got Environment) error {
		if got["TEST_VALUE"] != "preserved" {
			t.Fatalf("prepare environment = %#v", got)
		}
		return want
	})(env, nil)

	if !errors.Is(err, want) {
		t.Fatalf("Service() error = %v, want %v", err, want)
	}
}

func TestDBServiceRunsInitOnly(t *testing.T) {
	env := Environment{"TEST_VALUE": "preserved"}
	prepareCalled := false
	initCalled := false

	err := DBService(
		"component",
		func(Environment) error {
			prepareCalled = true
			return nil
		},
		func(got Environment) error {
			initCalled = true
			if got["TEST_VALUE"] != "preserved" {
				t.Fatalf("init environment = %#v", got)
			}
			return nil
		},
	)(env, []string{initDBCommand})

	if err != nil {
		t.Fatalf("DBService() error = %v", err)
	}
	if prepareCalled {
		t.Fatal("regular preparation ran for init-only command")
	}
	if !initCalled {
		t.Fatal("init-only preparation did not run")
	}
}

func TestStorageInitEntrypointRouting(t *testing.T) {
	for _, test := range []struct {
		args             []string
		handler, wantErr string
	}{
		{[]string{initHistoryProvidersCommand}, "history", ""},
		{[]string{initTelemetryCommand}, "telemetry", ""},
		{nil, "next", "startup failed"},
		{[]string{"zabbix_proxy"}, "next", "startup failed"},
		{[]string{"sh", "-c", "true"}, "next", "startup failed"},
		{[]string{initHistoryProvidersCommand, "extra"}, "", "takes no arguments"},
		{[]string{initTelemetryCommand, "extra"}, "", "takes no arguments"},
	} {
		t.Run(fmt.Sprint(test.args), func(t *testing.T) {
			env := Environment{"test": "value"}
			selected := ""
			record := func(name string, got Environment) {
				if selected != "" {
					t.Fatal("multiple handlers ran")
				}
				selected = name
				if !reflect.DeepEqual(got, env) {
					t.Fatal("environment changed")
				}
			}
			entry := StorageInitEntrypoint(func(got Environment, args []string) error {
				record("next", got)
				if !reflect.DeepEqual(args, test.args) {
					t.Fatal("command arguments changed")
				}
				return errors.New("startup failed")
			}, func(got Environment) (int, error) {
				record("history", got)
				return 1, nil
			}, func(got Environment) (int, error) {
				record("telemetry", got)
				return 1, nil
			})
			err := entry(env, test.args)
			if (err == nil) != (test.wantErr == "") || (err != nil && !strings.Contains(err.Error(), test.wantErr)) {
				t.Fatalf("error = %v, want %q", err, test.wantErr)
			}
			if selected != test.handler {
				t.Fatalf("handler = %q, want %q", selected, test.handler)
			}
		})
	}
}

func TestStorageInitEntrypointResult(t *testing.T) {
	for _, test := range []struct {
		name    string
		command string
		count   int
		err     error
		wantErr string
	}{
		{"prepared", initHistoryProvidersCommand, 1, nil, ""},
		{"no history selected", initHistoryProvidersCommand, 0, nil, "no history schemas initialized: set ZBX_HISTORYPROVIDER_<N>_SCHEMA"},
		{"no telemetry selected", initTelemetryCommand, 0, nil, "no telemetry schemas initialized: ZBX_TELEMETRYPROVIDER_0_SCHEMA is not set"},
		{"failed", initHistoryProvidersCommand, 0, errors.New("schema preparation failed"), ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			prepare := func(Environment) (int, error) { return test.count, test.err }
			entry := StorageInitEntrypoint(nil, prepare, prepare)
			err := entry(Environment{}, []string{test.command})
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("error = %v, want %q", err, test.wantErr)
				}
			} else if !errors.Is(err, test.err) {
				t.Fatalf("error = %v, want %v", err, test.err)
			}
		})
	}
}

func TestStorageInitEntrypointRejectsMissingHistoryHandler(t *testing.T) {
	entry := StorageInitEntrypoint(func(Environment, []string) error {
		t.Fatal("unsupported command passed to normal startup")
		return nil
	}, nil, func(Environment) (int, error) {
		t.Fatal("telemetry initialized for a history command")
		return 0, nil
	})
	if err := entry(Environment{}, []string{initHistoryProvidersCommand}); err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("unsupported history initialization: %v", err)
	}
}
