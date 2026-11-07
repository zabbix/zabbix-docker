//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestPrepareService(t *testing.T) {
	homeDir := t.TempDir()
	encryptionDirectory := filepath.Join(homeDir, "enc_internal")
	if err := os.Mkdir(encryptionDirectory, 0o700); err != nil {
		t.Fatal(err)
	}

	env := bootstrap.Environment{
		"ZABBIX_USER_HOME_DIR": homeDir,
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
}

func TestPrepareServiceKeepsExplicitAllowedIP(t *testing.T) {
	homeDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(homeDir, "enc_internal"), 0o700); err != nil {
		t.Fatal(err)
	}
	env := bootstrap.Environment{
		"ZABBIX_USER_HOME_DIR": homeDir,
		"ZBX_ALLOWEDIP":        "192.0.2.1",
		"ZBX_CLEAR_ENV":        "false",
	}

	if err := prepareService(env); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_ALLOWEDIP"] != "192.0.2.1" {
		t.Fatalf("unexpected allowed IP: %q", env["ZBX_ALLOWEDIP"])
	}
	if env["ZABBIX_USER_HOME_DIR"] != homeDir {
		t.Fatal("private environment was cleared despite ZBX_CLEAR_ENV=false")
	}
}
