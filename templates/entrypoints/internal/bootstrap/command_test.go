//go:build windows

package bootstrap

import (
	"errors"
	"os"
	"os/exec"
	"reflect"
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
	})(env, []string{"component"})

	if !errors.Is(err, want) {
		t.Fatalf("Service() error = %v, want %v", err, want)
	}
}
