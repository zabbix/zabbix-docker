package web

import (
	"os"
	"os/exec"
	"os/signal"
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
	for _, command := range children {
		command.Env = env.List()
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	return supervise(children)
}

func supervise(children []*exec.Cmd) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(signals)

	type childResult struct {
		command *exec.Cmd
		err     error
	}
	results := make(chan childResult, len(children))

	for index, command := range children {
		if err := command.Start(); err != nil {
			for _, started := range children[:index] {
				_ = started.Process.Signal(syscall.SIGTERM)
			}

			for range index {
				<-results
			}

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

	var timer *time.Timer
	var timeout <-chan time.Time

	if firstResult.command != nil {
		timer = time.NewTimer(childExitTimeout)
		timeout = timer.C
		defer timer.Stop()
	}

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

	return firstResult.err
}
