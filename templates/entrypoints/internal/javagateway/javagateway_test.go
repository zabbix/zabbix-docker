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
		"ZBX_LISTEN_IP":       "192.0.2.1",
		"ZBX_SERVER":          "192.0.2.0/24,zabbix.example.com",
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

func TestCommandPassesJavaOptionsThroughEnvironment(t *testing.T) {
	for _, test := range []struct {
		name   string
		jdk    string
		legacy string
		want   string
	}{
		{name: "empty"},
		{name: "jdk only", jdk: `-Xmx128m -Dname="value with spaces"`, want: `-Xmx128m -Dname="value with spaces"`},
		{name: "legacy only", legacy: `-Xmx128m -Dname="value with spaces"`, want: `-Xmx128m -Dname="value with spaces"`},
		{name: "both", jdk: `-Xms64m -Dname="jdk value"`, legacy: `-Xmx128m -Dname="legacy value"`, want: `-Xms64m -Dname="jdk value" -Xmx128m -Dname="legacy value"`},
		{name: "invalid quoting delegated to java", legacy: `-Dname="unterminated`, want: `-Dname="unterminated`},
	} {
		t.Run(test.name, func(t *testing.T) {
			env := bootstrap.Environment{
				"JDK_JAVA_OPTIONS": test.jdk,
				"ZBX_JAVA_OPTS":    test.legacy,
			}
			command := buildCommand(env, "/etc/zabbix/zabbix_java_gateway_logback.xml", nil)
			bootstrap.ClearPrivateEnv(env)
			if got := env["JDK_JAVA_OPTIONS"]; got != test.want {
				t.Fatalf("JDK_JAVA_OPTIONS = %q, want %q", got, test.want)
			}
			for _, arg := range command {
				if strings.HasPrefix(arg, "-X") || strings.HasPrefix(arg, "-Dname=") {
					t.Fatalf("Java environment option added to command arguments: %q", arg)
				}
			}
		})
	}
}
