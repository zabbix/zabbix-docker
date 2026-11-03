package main

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

const allowedMethodsTemplate = "# AllowedAPIMethods=\n"

func newPrepareServiceFixture(t *testing.T) (bootstrap.Environment, string) {
	t.Helper()

	configDir := t.TempDir()
	templatePath := filepath.Join(configDir, allowedMethodsConfigFile)
	if err := os.WriteFile(templatePath, []byte(allowedMethodsTemplate), 0o644); err != nil {
		t.Fatal(err)
	}

	runtimePath := filepath.Join(t.TempDir(), allowedMethodsConfigFile)
	originalRuntimePath := runtimeAllowedMethodsConfigPath
	runtimeAllowedMethodsConfigPath = runtimePath
	t.Cleanup(func() { runtimeAllowedMethodsConfigPath = originalRuntimePath })

	return bootstrap.Environment{"ZABBIX_CONF_DIR": configDir}, runtimePath
}

func TestPrepareService(t *testing.T) {
	env, runtimePath := newPrepareServiceFixture(t)
	templatePath := filepath.Join(env["ZABBIX_CONF_DIR"], allowedMethodsConfigFile)
	env["ZBX_ALLOWEDAPIMETHODS_0"] = "apiinfo.version"
	env["ZBX_ALLOWEDAPIMETHODS_1"] = "host.get"
	if err := prepareService(env); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(runtimePath)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(runtimePath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("runtime security configuration mode = %o, want 600", info.Mode().Perm())
	}
	want := allowedMethodsTemplate +
		"AllowedAPIMethods=${ZBX_ALLOWEDAPIMETHODS_0}\n" +
		"AllowedAPIMethods=${ZBX_ALLOWEDAPIMETHODS_1}\n"
	if string(data) != want {
		t.Fatalf("runtime security configuration:\n%s\nwant:\n%s", data, want)
	}

	source, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(source) != allowedMethodsTemplate {
		t.Fatalf("security configuration template was modified: %q", source)
	}

	for _, name := range []string{"ZBX_ALLOWEDAPIMETHODS_0", "ZBX_ALLOWEDAPIMETHODS_1"} {
		if env[name] == "" {
			t.Fatalf("%s was removed from the service environment", name)
		}
	}
}

func TestPrepareServiceWithoutAllowedMethods(t *testing.T) {
	env, runtimePath := newPrepareServiceFixture(t)
	env["ZABBIX_CONF_DIR"] = "/nonexistent"

	if err := prepareService(env); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(runtimePath); !os.IsNotExist(err) {
		t.Fatalf("runtime configuration was created without allowed methods: %v", err)
	}
	if _, exists := env["ZABBIX_CONF_DIR"]; exists {
		t.Error("private environment was not cleared")
	}
}

func TestPrepareServiceRejectsIncompleteIndexes(t *testing.T) {
	env, runtimePath := newPrepareServiceFixture(t)
	env["ZBX_ALLOWEDAPIMETHODS_1"] = "host.get"

	err := prepareService(env)
	if err == nil || !strings.Contains(err.Error(), "index 0 is missing") {
		t.Fatalf("prepareService() error = %v, want missing index error", err)
	}
	if _, err := os.Stat(runtimePath); !os.IsNotExist(err) {
		t.Fatalf("invalid indexes created runtime configuration: %v", err)
	}
}

func TestPrepareServiceRejectsEmptyAllowedMethod(t *testing.T) {
	env, runtimePath := newPrepareServiceFixture(t)
	env["ZBX_ALLOWEDAPIMETHODS_0"] = ""

	err := prepareService(env)
	if err == nil || !strings.Contains(err.Error(), "ZBX_ALLOWEDAPIMETHODS_0 must not be empty") {
		t.Fatalf("prepareService() error = %v, want empty value error", err)
	}
	if _, err := os.Stat(runtimePath); !os.IsNotExist(err) {
		t.Fatalf("empty allowed method created runtime configuration: %v", err)
	}
}

func TestPrepareLicenseAcceptance(t *testing.T) {
	for _, test := range []struct {
		name      string
		version   string
		configDir string
		wantErr   bool
	}{
		{name: "environment version", version: " 2026-09-29 "},
		{name: "custom config directory", version: "2026-09-29", configDir: "/custom/etc"},
		{name: "version validation is delegated to Adapter", version: "another-version"},
		{name: "missing version", wantErr: true},
		{name: "blank version", version: " \t", wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			command, err := prepareLicenseAcceptance(bootstrap.Environment{
				"ZBX_EULAVERSION": test.version,
				"ZABBIX_CONF_DIR": test.configDir,
			})
			if (err != nil) != test.wantErr {
				t.Fatalf("prepareLicenseAcceptance() error = %v, wantErr = %v", err, test.wantErr)
			}
			if test.wantErr {
				return
			}
			configDir := test.configDir
			if configDir == "" {
				configDir = "/etc/zabbix"
			}
			want := []string{
				bridgeAdapterBinary, "-c", filepath.Join(configDir, "zabbix_bridge_adapter.conf"),
				"license", "accept", "--non-interactive", "--version", strings.TrimSpace(test.version),
			}
			if !reflect.DeepEqual(command, want) {
				t.Errorf("command = %q, want %q", command, want)
			}
		})
	}
}

func TestLicenseAcceptanceLogs(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	originalStdout := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() {
		os.Stdout = originalStdout
		writer.Close()
		reader.Close()
	})

	_, err = prepareLicenseAcceptance(
		bootstrap.Environment{"ZBX_EULAVERSION": "2026-09-29"},
	)
	if err != nil {
		t.Fatal(err)
	}
	writer.Close()
	os.Stdout = originalStdout
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	for _, message := range []string{
		`Requesting non-interactive EULA acceptance for version "2026-09-29" supplied by ZBX_EULAVERSION`,
		"verify the bundled EULA version and write acceptance evidence",
		"Configuration files will not be rewritten",
		"Adapter service will not be started",
		`Executing: ["/usr/sbin/zabbix_bridge_adapter" "-c" "/etc/zabbix/zabbix_bridge_adapter.conf" "license" "accept" "--non-interactive" "--version" "2026-09-29"]`,
	} {
		if !strings.Contains(string(data), message) {
			t.Errorf("acceptance log missing %q:\n%s", message, data)
		}
	}
}

func TestEntrypointRequiresVersionBeforePreparingConfiguration(t *testing.T) {
	err := entrypoint(bootstrap.Environment{}, []string{acceptLicenseCommand})
	if err == nil || !strings.Contains(err.Error(), "set ZBX_EULAVERSION") {
		t.Fatalf("entrypoint() error = %v, want explicit EULA version requirement", err)
	}
}

func TestAcceptLicenseDoesNotPrepareRuntimeConfig(t *testing.T) {
	env, runtimePath := newPrepareServiceFixture(t)
	configDir := env["ZABBIX_CONF_DIR"]
	env["ZBX_EULAVERSION"] = "2026-09-29"
	env["ZBX_ALLOWEDAPIMETHODS_1"] = "host.get"
	command, err := prepareCommand(env, []string{acceptLicenseCommand})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		bridgeAdapterBinary, "-c", filepath.Join(configDir, "zabbix_bridge_adapter.conf"),
		"license", "accept", "--non-interactive", "--version", "2026-09-29",
	}
	if !reflect.DeepEqual(command, want) {
		t.Errorf("command = %q, want %q", command, want)
	}
	if _, err := os.Stat(runtimePath); !os.IsNotExist(err) {
		t.Errorf("accept-license created runtime configuration: %v", err)
	}
	if _, exists := env["ZABBIX_CONF_DIR"]; exists {
		t.Error("private environment was not cleared")
	}
}

func TestAcceptLicenseRejectsExtraArguments(t *testing.T) {
	_, err := prepareCommand(bootstrap.Environment{"ZBX_EULAVERSION": "2026-09-29"}, []string{acceptLicenseCommand, "--version", "another-version"})
	if err == nil || !strings.Contains(err.Error(), "usage: accept-license") {
		t.Fatalf("prepareCommand() error = %v, want usage error", err)
	}
}

func TestNativeLicenseCommandsWithoutAllowedMethods(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "native non-interactive accept is unchanged",
			args: []string{bridgeAdapterBinary, "-c", "/etc/zabbix/zabbix_bridge_adapter.conf", "license", "accept", "--non-interactive"},
			want: []string{bridgeAdapterBinary, "-c", "/etc/zabbix/zabbix_bridge_adapter.conf", "license", "accept", "--non-interactive"},
		},
		{
			name: "native explicit version is unchanged",
			args: []string{bridgeAdapterBinary, "license", "accept", "--non-interactive", "--version", "explicit-version"},
			want: []string{bridgeAdapterBinary, "license", "accept", "--non-interactive", "--version", "explicit-version"},
		},
		{
			name: "interactive accept",
			args: []string{bridgeAdapterBinary, "license", "accept"},
			want: []string{bridgeAdapterBinary, "license", "accept"},
		},
		{
			name: "show",
			args: []string{bridgeAdapterBinary, "license", "show"},
			want: []string{bridgeAdapterBinary, "license", "show"},
		},
		{
			name: "status",
			args: []string{bridgeAdapterBinary, "license", "status"},
			want: []string{bridgeAdapterBinary, "license", "status"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			env, runtimePath := newPrepareServiceFixture(t)
			env["ZBX_EULAVERSION"] = "2026-09-29"

			command, err := prepareCommand(env, test.args)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(command, test.want) {
				t.Errorf("command = %q, want %q", command, test.want)
			}
			if _, err := os.Stat(runtimePath); !os.IsNotExist(err) {
				t.Errorf("license command created runtime configuration: %v", err)
			}
			if _, exists := env["ZABBIX_CONF_DIR"]; exists {
				t.Error("private environment was not cleared")
			}
			if env["ZBX_EULAVERSION"] == "" {
				t.Error("EULA version was removed from the environment")
			}
		})
	}
}

func TestNativeCommandWithAllowedMethods(t *testing.T) {
	env, runtimePath := newPrepareServiceFixture(t)
	env["ZBX_ALLOWEDAPIMETHODS_0"] = "host.get"
	args := []string{bridgeAdapterBinary, "license", "show"}
	command, err := prepareCommand(env, args)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(command, args) {
		t.Fatalf("command = %q, want %q", command, args)
	}
	data, err := os.ReadFile(runtimePath)
	want := allowedMethodsTemplate + "AllowedAPIMethods=${ZBX_ALLOWEDAPIMETHODS_0}\n"
	if err != nil || string(data) != want {
		t.Fatalf("runtime configuration = %q, error = %v", data, err)
	}
}

func TestPrepareCommandForService(t *testing.T) {
	env, runtimePath := newPrepareServiceFixture(t)
	env["ZBX_EULAVERSION"] = "2026-09-29"
	env["ZBX_ALLOWEDAPIMETHODS_0"] = "apiinfo.version"
	command, err := prepareCommand(env, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(command, []string{bridgeAdapterBinary}) {
		t.Errorf("command = %q, want Adapter binary", command)
	}
	data, err := os.ReadFile(runtimePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "AllowedAPIMethods=${ZBX_ALLOWEDAPIMETHODS_0}") {
		t.Errorf("runtime allowed methods were not prepared: %s", data)
	}
}

func TestPrepareCommandForCustomCommand(t *testing.T) {
	env := bootstrap.Environment{"ZABBIX_CONF_DIR": "/nonexistent"}
	args := []string{"/bin/echo", "license", "accept", "--non-interactive"}
	command, err := prepareCommand(env, args)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(command, args) {
		t.Errorf("custom command was modified: %q", command)
	}
	if env["ZABBIX_CONF_DIR"] != "/nonexistent" {
		t.Error("custom command environment was modified")
	}
}

func TestPrepareServiceWithReadOnlyConfig(t *testing.T) {
	env, runtimePath := newPrepareServiceFixture(t)
	env["ZBX_ALLOWEDAPIMETHODS_0"] = "host.get"
	configDir := env["ZABBIX_CONF_DIR"]
	templatePath := filepath.Join(configDir, allowedMethodsConfigFile)
	if err := os.Chmod(templatePath, 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(configDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(configDir, 0o755); err != nil {
			t.Error(err)
		}
	})
	if err := prepareService(env); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(runtimePath)
	want := allowedMethodsTemplate + "AllowedAPIMethods=${ZBX_ALLOWEDAPIMETHODS_0}\n"
	if err != nil || string(data) != want {
		t.Fatalf("runtime configuration = %q, error = %v", data, err)
	}
	entries, err := os.ReadDir(configDir)
	if err != nil || len(entries) != 1 || entries[0].Name() != allowedMethodsConfigFile {
		t.Fatalf("configuration directory was modified: %v, error = %v", entries, err)
	}
	data, err = os.ReadFile(templatePath)
	if err != nil || string(data) != allowedMethodsTemplate {
		t.Errorf("configuration template was modified: %q, error = %v", data, err)
	}
}
