package web

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

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

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(signals)

	if err := php.Start(); err != nil {
		return err
	}

	if err := web.Start(); err != nil {
		_ = php.Process.Signal(syscall.SIGTERM)
		_ = php.Wait()
		return err
	}

	type result struct {
		command *exec.Cmd
		err     error
	}
	results := make(chan result, len(children))

	go func() { results <- result{php, php.Wait()} }()
	go func() { results <- result{web, web.Wait()} }()

	// One child dying (or a stop signal) takes the whole stack down.
	remaining := len(children)
	var first result
	select {
	case first = <-results:
		remaining--
	case <-signals:
	}

	for _, command := range children {
		if command != first.command && command.Process != nil {
			_ = command.Process.Signal(syscall.SIGTERM)
		}
	}

	for ; remaining > 0; remaining-- {
		<-results
	}

	return first.err
}
