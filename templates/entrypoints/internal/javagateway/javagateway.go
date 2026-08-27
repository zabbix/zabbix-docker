// Package javagateway prepares and starts the Zabbix Java Gateway.
package javagateway

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattn/go-shellwords"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

const (
	serviceCommand = "java-gateway"
	javaDir        = "/usr/sbin/zabbix_java"
	javaClasspath  = "lib/*:bin/*:ext_lib/*"
	javaMainClass  = "com.zabbix.gateway.JavaGateway"
	logConfigName  = "zabbix_java_gateway_logback.xml"
)

// Run implements the Java Gateway entrypoint. Empty arguments, the
// java-gateway pseudo-command and options start the service; any other command
// is executed unchanged.
func Run(env bootstrap.Environment, args []string) error {
	extraArgs, startsService := bootstrap.ServiceArgs(args, serviceCommand)
	if !startsService {
		return bootstrap.Execute(args, env)
	}

	command, err := prepare(env, extraArgs)
	if err != nil {
		return err
	}

	return bootstrap.Execute(command, env)
}

func prepare(env bootstrap.Environment, extraArgs []string) ([]string, error) {
	bootstrap.LogInfo("** Preparing Zabbix Java Gateway")

	configDir, err := bootstrap.RequiredConfigDir(env)
	if err != nil {
		return nil, err
	}

	env.SetDefaultNonEmpty("ZBX_TIMEOUT", "3")
	env.SetDefaultNonEmpty("ZBX_DEBUGLEVEL", "info")

	logConfig := filepath.Join(configDir, logConfigName)
	if !bootstrap.RegularFile(logConfig) {
		return nil, fmt.Errorf("missing configuration file %s", logConfig)
	}

	if err := os.Chdir(javaDir); err != nil {
		return nil, fmt.Errorf("change Java Gateway directory to %s: %w", javaDir, err)
	}

	command, err := buildCommand(env, logConfig, extraArgs)
	if err != nil {
		return nil, err
	}

	bootstrap.ClearPrivateEnv(env)

	return command, nil
}

func buildCommand(env bootstrap.Environment, logConfig string, extraArgs []string) ([]string, error) {
	javaOpts := []string{
		"-server",
		"-Dlogback.configurationFile=" + logConfig,
	}

	parser := shellwords.NewParser()
	parser.ParseEnv = false
	parser.ParseBacktick = false

	extraJavaOpts, err := parser.Parse(env["ZBX_JAVA_OPTS"])
	if err != nil {
		return nil, fmt.Errorf("parse ZBX_JAVA_OPTS: %w", err)
	}

	javaOpts = append(javaOpts, extraJavaOpts...)
	javaOpts = append(javaOpts, extraArgs...)

	zabbixOpts := []string{
		"-Dsun.rmi.transport.tcp.responseTimeout=" + env["ZBX_TIMEOUT"] + "000",
		"-Dzabbix.listenPort=" + env.ValueOrDefaultNonEmpty("ZBX_LISTEN_PORT", "10052"),
		"-Dzabbix.timeout=" + env["ZBX_TIMEOUT"],
	}

	if value := env["ZBX_LISTEN_IP"]; value != "" {
		zabbixOpts = append(zabbixOpts, "-Dzabbix.listenIP="+value)
	}
	if value := env["ZBX_SERVER"]; value != "" {
		zabbixOpts = append(zabbixOpts, "-Dzabbix.server="+value)
	}
	if value := env["ZBX_START_POLLERS"]; value != "" {
		zabbixOpts = append(zabbixOpts, "-Dzabbix.startPollers="+value)
	}
	if value := env["ZBX_PROPERTIES_FILE"]; value != "" {
		zabbixOpts = append(zabbixOpts, "-Dzabbix.propertiesFile="+value)
	}

	command := []string{env.ValueOrDefaultNonEmpty("JAVA", "/usr/bin/java")}
	command = append(command, javaOpts...)
	command = append(command, "-classpath", javaClasspath)
	command = append(command, zabbixOpts...)

	return append(command, javaMainClass), nil
}
