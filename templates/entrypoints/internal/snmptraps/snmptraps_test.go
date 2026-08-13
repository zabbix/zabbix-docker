package snmptraps

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestBuildCommand(t *testing.T) {
	persistentDir := t.TempDir()
	for _, name := range []string{"snmptrapd.conf", "snmptrapd_custom.conf"} {
		if err := os.WriteFile(filepath.Join(persistentDir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	env := bootstrap.Environment{
		"DEBUG_MODE":              "true",
		"SNMP_PERSISTENT_DIR":     persistentDir,
		"SNMPTRAP_OUTPUT_OPTIONS": "nq",
	}

	got := buildCommand(env, []string{"--debug"})
	want := []string{
		snmptrapdBinary,
		"-f",
		"-a",
		"-C",
		"-c", "/etc/snmp/snmptrapd.conf," +
			filepath.Join(persistentDir, "snmptrapd.conf") + "," +
			filepath.Join(persistentDir, "snmptrapd_custom.conf"),
		"-t",
		"-X",
		"-Lo",
		"--hexOutputLength=0",
		"-Onq",
		"-n",
		"-DALL",
		"--debug",
		"udp:1162",
		"udp6:1162",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildCommand() = %#v, want %#v", got, want)
	}
}

func TestBuildCommandDefaults(t *testing.T) {
	got := buildCommand(bootstrap.Environment{
		"SNMPTRAP_OUTPUT_OPTIONS": defaultOutputOptions,
	}, nil)
	if got[4] != "-c" || got[5] != "/etc/snmp/snmptrapd.conf" {
		t.Fatalf("unexpected config arguments: %q", got)
	}
	if !slices.Contains(got, "-O"+defaultOutputOptions) {
		t.Fatalf("unexpected output options: %q", got)
	}
}
