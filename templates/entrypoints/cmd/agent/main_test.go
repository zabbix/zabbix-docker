//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestPrepareService(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "conf")
	homeDir := filepath.Join(root, "home")
	if err := os.MkdirAll(filepath.Join(homeDir, "enc_internal"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(homeDir, "enc"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "zabbix_agentd.conf")
	config := "# Server=\n# ServerActive=\n# LogType=file\nLogFile=C:\\zabbix_agentd.log\nLogFileSize=1\n# ListenBacklog=\n# Hostname=\n# Include=\n# TLSPSKIdentity=\n# TLSPSKFile=\n# DenyKey=system.run[*]\n# AllowKey=\n"
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}

	env := bootstrap.Environment{
		"ZABBIX_CONF_DIR": configDir, "ZABBIX_USER_HOME_DIR": homeDir,
		"ZBX_HOSTNAME": "agent-host", "ZBX_SOURCEIP": "192.0.2.1", "ZBX_LISTENBACKLOG": "128", "ZBX_TLSPSK": "secret", "ZBX_TLSCAFILE": "ca.pem",
		"ZBX_TLSPSKIDENTITY": "identity", "ZBX_DENYKEY": "system.run[*]",
		"ZBX_ALLOWKEY":       "system.localtime",
		"UNRELATED_VARIABLE": "value",
	}
	if err := prepareService(env); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	wants := []string{
		"Server=zabbix-server\n",
		"ServerActive=zabbix-server\n",
		"LogType=console\n",
		"Hostname=agent-host\n",
		"SourceIP=192.0.2.1\n",
		"ListenBacklog=128\n",
		"Include=" + filepath.Join(configDir, "zabbix_agentd.d", "*.conf") + "\n",
		"TLSPSKIdentity=identity\n",
		"TLSPSKFile=" + filepath.Join(homeDir, "enc_internal", "ZBX_TLSPSKFILE") + "\n",
		"TLSCAFile=" + filepath.Join(homeDir, "enc", "ca.pem") + "\n",
		"DenyKey=system.run[*]\n",
		"AllowKey=system.localtime\n",
	}
	for _, want := range wants {
		if !strings.Contains(content, want) {
			t.Fatalf("configuration does not contain %q:\n%s", want, data)
		}
	}
	if strings.Contains("\n"+content, "\nLogFile=") || strings.Contains("\n"+content, "\nLogFileSize=") {
		t.Fatalf("file logging parameters were not removed:\n%s", data)
	}

	tlsData, err := os.ReadFile(filepath.Join(homeDir, "enc_internal", "ZBX_TLSPSKFILE"))
	if err != nil {
		t.Fatal(err)
	}
	if string(tlsData) != "secret" {
		t.Fatalf("unexpected TLS PSK: %q", tlsData)
	}
	if env["UNRELATED_VARIABLE"] != "value" {
		t.Fatal("entrypoint removed an unrelated variable")
	}
	if _, found := env["ZBX_HOSTNAME"]; found {
		t.Fatal("ZBX_HOSTNAME was not removed")
	}
	if env["ZABBIX_CONF_DIR"] != configDir {
		t.Fatal("ZABBIX_CONF_DIR was unexpectedly removed")
	}
}
