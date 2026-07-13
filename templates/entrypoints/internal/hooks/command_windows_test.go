//go:build windows

package hooks

import (
	"reflect"
	"testing"
)

func TestPowerShellCommand(t *testing.T) {
	path := `C:\zabbix\entrypoint.d\10-custom.ps1`
	args, supported := command(path)
	want := []string{"pwsh.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", path}
	if !supported || !reflect.DeepEqual(args, want) {
		t.Fatalf("command() = %#v, %v; want %#v, true", args, supported, want)
	}
}

func TestCmdCommand(t *testing.T) {
	path := `C:\zabbix\entrypoint.d\20-custom.cmd`
	args, supported := command(path)
	want := []string{"cmd.exe", "/D", "/S", "/C", path}
	if !supported || !reflect.DeepEqual(args, want) {
		t.Fatalf("command() = %#v, %v; want %#v, true", args, supported, want)
	}
}
