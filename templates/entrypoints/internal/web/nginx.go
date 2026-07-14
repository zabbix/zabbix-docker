package web

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func prepareNginx(env bootstrap.Environment) error {
	includes := env.ValueOrDefaultNonEmpty("NGINX_INCLUDES_DIR", "/etc/nginx/includes")

	for _, name := range []string{"access-log.conf", "user.conf", "real-ip.conf", "listen-ipv6.conf", "listen-ipv6-ssl.conf"} {
		if err := os.WriteFile(filepath.Join(includes, name), nil, 0o644); err != nil {
			return err
		}
	}

	if os.Getuid() == 0 {
		if err := os.WriteFile(filepath.Join(includes, "user.conf"), []byte(fmt.Sprintf("user %s;\n", env["DAEMON_USER"])), 0o644); err != nil {
			return err
		}
	}

	configDir := env.ValueOrDefaultNonEmpty("ZABBIX_CONF_DIR", "/etc/zabbix")
	confDirectory := env.ValueOrDefaultNonEmpty("NGINX_CONFD_DIR", "/etc/nginx/http.d")

	bootstrap.LogInfo("** Adding Zabbix virtual host (HTTP)")
	if err := bootstrap.ReplaceSymlink(filepath.Join(configDir, "nginx.conf"), filepath.Join(confDirectory, "nginx.conf")); err != nil {
		if !errors.Is(err, bootstrap.ErrSymlinkSourceNotExist) {
			return fmt.Errorf("enable Nginx HTTP virtual host: %w", err)
		}
		bootstrap.LogWarn("**** Impossible to enable HTTP virtual host")
	} else if ipv6Available() {
		if err := os.WriteFile(filepath.Join(includes, "listen-ipv6.conf"), []byte("listen [::]:8080;\nallow ::1;\n"), 0o644); err != nil {
			return err
		}
	}

	sslDirectory := env.ValueOrDefaultNonEmpty("NGINX_SSL_CONFIG_DIR", "/etc/ssl/nginx")
	if bootstrap.RegularFile(filepath.Join(sslDirectory, "ssl.crt")) && bootstrap.RegularFile(filepath.Join(sslDirectory, "ssl.key")) && bootstrap.RegularFile(filepath.Join(sslDirectory, "dhparam.pem")) {
		bootstrap.LogInfo("** Enable SSL support for Nginx")
		if err := bootstrap.ReplaceSymlink(filepath.Join(configDir, "nginx_ssl.conf"), filepath.Join(confDirectory, "nginx_ssl.conf")); err != nil {
			if !errors.Is(err, bootstrap.ErrSymlinkSourceNotExist) {
				return fmt.Errorf("enable Nginx HTTPS virtual host: %w", err)
			}
			bootstrap.LogWarn("**** Impossible to enable HTTPS virtual host")
		} else if ipv6Available() {
			if err := os.WriteFile(filepath.Join(includes, "listen-ipv6-ssl.conf"), []byte("listen [::]:8443 ssl;\nallow ::1;\n"), 0o644); err != nil {
				return err
			}
		}
	} else {
		bootstrap.LogWarn("**** Impossible to enable SSL support for Nginx. Certificates are missing.")
	}

	maxExecutionTime, err := strconv.Atoi(env.ValueOrDefaultNonEmpty("ZBX_MAXEXECUTIONTIME", "3"))
	if err != nil {
		return fmt.Errorf("invalid ZBX_MAXEXECUTIONTIME: %w", err)
	}

	serverCommon := filepath.Join(includes, "server-common.conf")
	if err := bootstrap.ReplaceInFile(serverCommon, map[string]string{
		"{FCGI_READ_TIMEOUT}": strconv.Itoa(maxExecutionTime + 1),
		"{HTTP_INDEX_FILE}":   env.ValueOrDefaultNonEmpty("HTTP_INDEX_FILE", "index.php"),
	}); err != nil {
		return err
	}

	if err := bootstrap.ReplaceInFile(env.ValueOrDefaultNonEmpty("NGINX_CONF_FILE", "/etc/nginx/nginx.conf"), map[string]string{
		"{EXPOSE_WEB_SERVER_INFO}": env["EXPOSE_WEB_SERVER_INFO"],
	}); err != nil {
		return err
	}

	accessLog := "access_log /var/log/nginx/access.log main;\n"
	if strings.EqualFold(env.ValueOrDefaultNonEmpty("ENABLE_WEB_ACCESS_LOG", "true"), "false") {
		accessLog = "access_log off;\n"
	}
	if err := os.WriteFile(filepath.Join(includes, "access-log.conf"), []byte(accessLog), 0o644); err != nil {
		return err
	}

	if from := env["WEB_REAL_IP_FROM"]; from != "" {
		content := fmt.Sprintf("set_real_ip_from %s;\n", from)
		if header := env["WEB_REAL_IP_HEADER"]; header != "" {
			content += fmt.Sprintf("real_ip_header %s;\n", header)
		}
		return os.WriteFile(filepath.Join(includes, "real-ip.conf"), []byte(content), 0o644)
	}

	return nil
}

func ipv6Available() bool {
	listener, err := net.Listen("tcp6", "[::]:0")
	if err != nil {
		return false
	}
	_ = listener.Close()
	return true
}
