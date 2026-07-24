package snmptraps

import (
	"os"
	"path/filepath"
	"reflect"
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
		"SNMP_PERSISTENT_DIR":     persistentDir,
		"SNMPTRAP_OUTPUT_OPTIONS": "nq",
	}

	got := buildCommand(env, []string{"--debug"})
	want := []string{
		snmptrapdBinary,
		"--doNotFork=yes",
		"-C",
		"-c", "/etc/snmp/snmptrapd.conf," +
			filepath.Join(persistentDir, "snmptrapd.conf") + "," +
			filepath.Join(persistentDir, "snmptrapd_custom.conf"),
		"-n",
		"-t",
		"-X",
		"-Lo",
		"-A",
		"-Onq",
		"--debug",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildCommand() = %#v, want %#v", got, want)
	}
}

func TestBuildCommandDefaults(t *testing.T) {
	got := buildCommand(bootstrap.Environment{
		"SNMPTRAP_OUTPUT_OPTIONS": defaultOutputOptions,
	}, nil)
	if got[3] != "-c" || got[4] != "/etc/snmp/snmptrapd.conf" {
		t.Fatalf("unexpected config arguments: %q", got)
	}
	if got[len(got)-1] != "-OSTte" {
		t.Fatalf("unexpected output options: %q", got)
	}
}
