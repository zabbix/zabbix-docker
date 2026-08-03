package javagateway

import (
	"reflect"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

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
		"ZBX_JAVA_OPTS":       `-Xms64m -Xmx128m -Dname="value with spaces"`,
		"ZBX_LISTEN_IP":       "192.0.2.1",
		"ZBX_SERVER":          "192.0.2.0/24,zabbix.example.com",
		"ZBX_START_POLLERS":   "7",
		"ZBX_PROPERTIES_FILE": "/tmp/gateway.properties",
	}

	got, err := buildCommand(
		env,
		"/etc/zabbix/zabbix_java_gateway_logback.xml",
		[]string{"-Dcustom=true"},
	)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"/custom/java",
		"-server",
		"-Dlogback.configurationFile=/etc/zabbix/zabbix_java_gateway_logback.xml",
		"-Xms64m",
		"-Xmx128m",
		"-Dname=value with spaces",
		"-Dcustom=true",
		"-classpath",
		"lib/*:bin/*:ext_lib/*",
		"-Dsun.rmi.transport.tcp.responseTimeout=5000",
		"-Dzabbix.listenPort=10053",
		"-Dzabbix.timeout=5",
		"-Dzabbix.listenIP=192.0.2.1",
		"-Dzabbix.server=192.0.2.0/24,zabbix.example.com",
		"-Dzabbix.startPollers=7",
		"-Dzabbix.propertiesFile=/tmp/gateway.properties",
		"com.zabbix.gateway.JavaGateway",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildCommand() = %#v, want %#v", got, want)
	}
}

func TestCommandRejectsInvalidJavaOptions(t *testing.T) {
	_, err := buildCommand(
		bootstrap.Environment{"ZBX_JAVA_OPTS": `-Dname="unterminated`},
		"/etc/zabbix/zabbix_java_gateway_logback.xml",
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "parse ZBX_JAVA_OPTS") {
		t.Fatalf("buildCommand() error = %v, want parsing error", err)
	}
}
