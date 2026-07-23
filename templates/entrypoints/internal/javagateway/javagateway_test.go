package javagateway

import (
	"reflect"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestServiceArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		want       []string
		wantStarts bool
	}{
		{name: "empty", wantStarts: true},
		{name: "service command", args: []string{"java-gateway"}, want: []string{}, wantStarts: true},
		{
			name:       "service options",
			args:       []string{"java-gateway", "-Dcustom=true"},
			want:       []string{"-Dcustom=true"},
			wantStarts: true,
		},
		{name: "options", args: []string{"--version"}, want: []string{"--version"}, wantStarts: true},
		{name: "custom command", args: []string{"sh"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, starts := serviceArgs(test.args)
			if starts != test.wantStarts || !reflect.DeepEqual(got, test.want) {
				t.Fatalf("serviceArgs(%q) = %q, %t; want %q, %t",
					test.args, got, starts, test.want, test.wantStarts)
			}
		})
	}
}

func TestPrepareRequiresLogConfig(t *testing.T) {
	env := bootstrap.Environment{
		"ZABBIX_USER_HOME_DIR": t.TempDir(),
		"ZABBIX_CONF_DIR":      t.TempDir(),
	}

	_, err := prepare(env, nil)
	if err == nil || !strings.Contains(err.Error(), "missing configuration file") {
		t.Fatalf("prepare() error = %v, want missing configuration file error", err)
	}
}

func TestCommandOptions(t *testing.T) {
	env := bootstrap.Environment{
		"JAVA":                "/custom/java",
		"ZBX_TIMEOUT":         "5",
		"ZBX_DEBUGLEVEL":      "debug",
		"ZBX_LISTEN_PORT":     "10053",
		"ZBX_JAVA_OPTS":       "-Xms64m -Xmx128m",
		"ZBX_LISTEN_IP":       "192.0.2.1",
		"ZBX_START_POLLERS":   "7",
		"ZBX_PROPERTIES_FILE": "/tmp/gateway.properties",
	}

	got := buildCommand(
		env,
		"/etc/zabbix/zabbix_java_gateway_logback.xml",
		[]string{"-Dcustom=true"},
	)
	want := []string{
		"/custom/java",
		"-server",
		"-Dlogback.configurationFile=/etc/zabbix/zabbix_java_gateway_logback.xml",
		"-Xms64m",
		"-Xmx128m",
		"-Dcustom=true",
		"-classpath",
		"lib/*:bin/*:ext_lib/*",
		"-Dsun.rmi.transport.tcp.responseTimeout=5000",
		"-Dzabbix.listenPort=10053",
		"-Dzabbix.timeout=5",
		"-Dzabbix.pidFile=/tmp/java_gateway.pid",
		"-Dzabbix.listenIP=192.0.2.1",
		"-Dzabbix.startPollers=7",
		"-Dzabbix.propertiesFile=/tmp/gateway.properties",
		"com.zabbix.gateway.JavaGateway",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildCommand() = %#v, want %#v", got, want)
	}
}
