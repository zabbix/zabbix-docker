package web

import (
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

const childExitTimeout = 10 * time.Second

func startStack(env bootstrap.Environment, server ServerType) error {
	phpBinary := env.ValueOrDefaultNonEmpty("PHP_FPM_BIN", "/usr/sbin/php-fpm")
	phpConfig := env.ValueOrDefaultNonEmpty("PHP_FPM_CONFIG", "/etc/php-fpm.conf")

	php := exec.Command(phpBinary, "--nodaemonize", "--fpm-config", phpConfig)

	web := exec.Command(env.ValueOrDefaultNonEmpty("NGINX_BIN", "/usr/sbin/nginx"), "-e", "stderr", "-g", "daemon off;", "-c", env.ValueOrDefaultNonEmpty("NGINX_CONF_FILE", "/etc/nginx/nginx.conf"))
	if server == Apache {
		web = exec.Command(env.ValueOrDefaultNonEmpty("APACHE_BIN", "/usr/sbin/httpd"), "-D", "FOREGROUND")
	}

	children := []*exec.Cmd{php, web}
	php.Env = env.List()
	web.Env = webServerEnv(env).List()
	for _, command := range children {
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	return supervise(children)
}

func webServerEnv(env bootstrap.Environment) bootstrap.Environment {
	result := make(bootstrap.Environment, len(env))
	for name, value := range env {
		result[name] = value
	}

	for name := range result {
		if name == "VAULT_TOKEN" {
			delete(result, name)
			continue
		}

		for _, prefix := range []string{"DB_", "MYSQL_", "POSTGRES_", "ZBX_DB_", "ZBX_VAULT"} {
			if strings.HasPrefix(name, prefix) {
				delete(result, name)
				break
			}
		}
	}

	return result
}

type childResult struct {
	command *exec.Cmd
	err     error
}

func supervise(children []*exec.Cmd) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(signals)

	results := make(chan childResult, len(children))

	for index, command := range children {
		if err := command.Start(); err != nil {
			for _, started := range children[:index] {
				_ = started.Process.Signal(syscall.SIGTERM)
			}

			waitForChildren(children[:index], results, index)

			return err
		}

		go func(command *exec.Cmd) {
			results <- childResult{command, command.Wait()}
		}(command)
	}

	remaining := len(children)
	var firstResult childResult

	select {
	case firstResult = <-results:
		remaining--
	case <-signals:
	}

	for _, command := range children {
		if command != firstResult.command && command.Process != nil {
			_ = command.Process.Signal(syscall.SIGTERM)
		}
	}

	waitForChildren(children, results, remaining)

	return firstResult.err
}

func waitForChildren(children []*exec.Cmd, results <-chan childResult, remaining int) {
	if remaining == 0 {
		return
	}

	timer := time.NewTimer(childExitTimeout)
	defer timer.Stop()
	timeout := timer.C

	for remaining > 0 {
		select {
		case <-results:
			remaining--
		case <-timeout:
			bootstrap.LogWarn("** Web stack did not stop within %s; forcing shutdown", childExitTimeout)

			for _, command := range children {
				if command.Process != nil {
					_ = command.Process.Kill()
				}
			}

			timeout = nil
		}
	}
}
