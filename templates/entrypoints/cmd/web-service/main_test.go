//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestPrepareService(t *testing.T) {
	homeDirectory := t.TempDir()
	encryptionDirectory := filepath.Join(homeDirectory, "enc_internal")
	if err := os.Mkdir(encryptionDirectory, 0o700); err != nil {
		t.Fatal(err)
	}

	env := bootstrap.Environment{
		"ZABBIX_USER_HOME_DIR": homeDirectory,
		"ZBX_TLSCA":            "ca-data",
		"ZBX_TLSCERT":          "certificate-data",
		"ZBX_TLSKEY":           "key-data",
		"MYSQL_PASSWORD":       "password",
	}
	if err := prepareService(env); err != nil {
		t.Fatal(err)
	}

	if env["ZBX_ALLOWEDIP"] != "zabbix-server" {
		t.Fatalf("unexpected allowed IP: %q", env["ZBX_ALLOWEDIP"])
	}
	if _, found := env["MYSQL_PASSWORD"]; found {
		t.Fatal("MYSQL_PASSWORD was not removed")
	}
	for _, variable := range []string{"ZBX_TLSCA", "ZBX_TLSCERT", "ZBX_TLSKEY"} {
		if _, found := env[variable]; found {
			t.Fatalf("%s was not removed", variable)
		}
	}
	for variable, expected := range map[string]string{
		"ZBX_TLSCAFILE":   "ca-data",
		"ZBX_TLSCERTFILE": "certificate-data",
		"ZBX_TLSKEYFILE":  "key-data",
	} {
		path := env[variable]
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if string(data) != expected {
			t.Fatalf("unexpected %s content: %q", variable, data)
		}
	}
}

func TestPrepareServiceKeepsExplicitAllowedIP(t *testing.T) {
	homeDirectory := t.TempDir()
	if err := os.Mkdir(filepath.Join(homeDirectory, "enc_internal"), 0o700); err != nil {
		t.Fatal(err)
	}
	env := bootstrap.Environment{
		"ZABBIX_USER_HOME_DIR": homeDirectory,
		"ZBX_ALLOWEDIP":        "192.0.2.1",
		"ZBX_CLEAR_ENV":        "false",
	}

	if err := prepareService(env); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_ALLOWEDIP"] != "192.0.2.1" {
		t.Fatalf("unexpected allowed IP: %q", env["ZBX_ALLOWEDIP"])
	}
	if env["ZABBIX_USER_HOME_DIR"] != homeDirectory {
		t.Fatal("private environment was cleared despite ZBX_CLEAR_ENV=false")
	}
}
