//go:build windows

package hooks

import (
	"os"
	"reflect"
	"testing"
)

func TestPowerShellCommand(t *testing.T) {
	path := `C:\zabbix\conf\entrypoint.d\10-custom.ps1`
	args, supported := hookCommand(path, os.FileMode(0))
	want := []string{"pwsh.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", path}
	if !supported || !reflect.DeepEqual(args, want) {
		t.Fatalf("hookCommand() = %#v, %v; want %#v, true", args, supported, want)
	}
}

func TestCmdCommand(t *testing.T) {
	path := `C:\zabbix\conf\entrypoint.d\20-custom.cmd`
	args, supported := hookCommand(path, os.FileMode(0))
	want := []string{"cmd.exe", "/D", "/S", "/C", path}
	if !supported || !reflect.DeepEqual(args, want) {
		t.Fatalf("hookCommand() = %#v, %v; want %#v, true", args, supported, want)
	}
}
