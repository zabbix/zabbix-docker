package bootstrap

import (
	"errors"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func TestCommand(t *testing.T) {
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
			if got := Command(test.args, "component"); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Command() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestExitCode(t *testing.T) {
	if code := ExitCode(nil); code != 0 {
		t.Fatalf("ExitCode(nil) = %d, want 0", code)
	}
	if code := ExitCode(errors.New("failure")); code != 1 {
		t.Fatalf("ExitCode(regular error) = %d, want 1", code)
	}

	command := exec.Command(os.Args[0], "-test.run=TestExitCodeHelperProcess")
	command.Env = append(os.Environ(), "ENTRYPOINT_EXIT_CODE_HELPER=1")
	err := command.Run()
	if code := ExitCode(err); code != 23 {
		t.Fatalf("ExitCode(child error) = %d, want 23: %v", code, err)
	}
}

func TestExitCodeHelperProcess(t *testing.T) {
	if os.Getenv("ENTRYPOINT_EXIT_CODE_HELPER") != "1" {
		return
	}

	os.Exit(23)
}
