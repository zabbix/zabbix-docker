package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestPrepareExportDir(t *testing.T) {
	tests := []struct {
		name        string
		size        string
		setSize     bool
		exportDir   string
		wantPresent bool
	}{
		{name: "size unset", exportDir: "/var/lib/zabbix/export/"},
		{name: "size empty", setSize: true, exportDir: "/var/lib/zabbix/export/"},
		{name: "size set", setSize: true, size: "1G", exportDir: "/var/lib/zabbix/export/", wantPresent: true},
		{name: "custom directory", setSize: true, size: "100M", exportDir: "/custom/export", wantPresent: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			homeDir := t.TempDir()
			configDir := t.TempDir()
			caDir := filepath.Join(homeDir, "ssl", "ssl_ca")
			internalCADir := filepath.Join(homeDir, "ssl", "ssl_ca_internal")
			for _, dir := range []string{caDir, internalCADir} {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(configDir, "zabbix_server_modules.conf"), []byte("# LoadModule=\n"), 0o644); err != nil {
				t.Fatal(err)
			}

			env := bootstrap.Environment{
				"ZABBIX_USER_HOME_DIR": homeDir,
				"ZABBIX_CONF_DIR":      configDir,
				"ZBX_SSLCALOCATION":    internalCADir,
				"ZBX_EXPORTDIR":        test.exportDir,
			}
			if test.setSize {
				env["ZBX_EXPORTFILESIZE"] = test.size
			}

			if err := Prepare(env); err != nil {
				t.Fatal(err)
			}
			gotDir, present := env["ZBX_EXPORTDIR"]
			if present != test.wantPresent || (present && gotDir != test.exportDir) {
				t.Fatalf("ZBX_EXPORTDIR = %q (present: %t), want %q (present: %t)", gotDir, present, test.exportDir, test.wantPresent)
			}
		})
	}
}
