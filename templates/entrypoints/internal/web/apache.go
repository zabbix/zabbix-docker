package web

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func prepareApache(env bootstrap.Environment) error {
	if os.Getuid() == 0 {
		env["APACHE_RUN_USER"] = env["DAEMON_USER"]
	} else {
		account, err := user.LookupId(strconv.Itoa(os.Getuid()))
		if err != nil {
			return err
		}
		env["APACHE_RUN_USER"] = account.Username
	}
	env["APACHE_RUN_GROUP"] = env["DAEMON_GROUP"]

	configDir := env.ValueOrDefaultNonEmpty("ZABBIX_CONF_DIR", "/etc/zabbix")
	sitesDirectory := env.ValueOrDefaultNonEmpty("APACHE_SITES_DIR", "/etc/apache2/conf.d")

	bootstrap.LogInfo("** Adding Zabbix virtual host (HTTP)")
	if err := bootstrap.ReplaceSymlink(filepath.Join(configDir, "apache.conf"), filepath.Join(sitesDirectory, "zabbix.conf")); err != nil {
		if !errors.Is(err, bootstrap.ErrSymlinkSourceNotExist) {
			return fmt.Errorf("enable Apache HTTP virtual host: %w", err)
		}
		bootstrap.LogWarn("**** Impossible to enable HTTP virtual host")
	}

	sslDirectory := env.ValueOrDefaultNonEmpty("APACHE_SSL_CONFIG_DIR", "/etc/ssl/apache2")
	if bootstrap.RegularFile(filepath.Join(sslDirectory, "ssl.crt")) && bootstrap.RegularFile(filepath.Join(sslDirectory, "ssl.key")) {
		bootstrap.LogInfo("** Adding Zabbix virtual host (HTTPS)")
		if err := bootstrap.ReplaceSymlink(filepath.Join(configDir, "apache_ssl.conf"), filepath.Join(sitesDirectory, "zabbix_ssl.conf")); err != nil {
			if !errors.Is(err, bootstrap.ErrSymlinkSourceNotExist) {
				return fmt.Errorf("enable Apache HTTPS virtual host: %w", err)
			}
			bootstrap.LogWarn("**** Impossible to enable HTTPS virtual host")
		}
	} else {
		bootstrap.LogWarn("**** Impossible to enable SSL support for Apache2. Certificates are missing.")
	}

	env.SetDefaultNonEmpty("HTTP_INDEX_FILE", "index.php")

	env["APACHE_CUSTOM_LOG"] = "/proc/self/fd/1"
	if strings.EqualFold(env.ValueOrDefaultNonEmpty("ENABLE_WEB_ACCESS_LOG", "true"), "false") {
		env["APACHE_CUSTOM_LOG"] = "/dev/null"
	}

	env["APACHE_SERVER_TOKENS"] = "OS"
	env["APACHE_SERVER_SIGNATURE"] = "On"
	if env["EXPOSE_WEB_SERVER_INFO"] == "off" {
		env["APACHE_SERVER_TOKENS"] = "Prod"
		env["APACHE_SERVER_SIGNATURE"] = "Off"
	}

	commonConfig := filepath.Join(sitesDirectory, "server-common.inc")
	if env["WEB_REAL_IP_FROM"] == "" {
		if err := bootstrap.RemoveLinesContaining(commonConfig, "WEB_REAL_IP_FROM"); err != nil {
			return err
		}
	}
	if env["WEB_REAL_IP_HEADER"] == "" {
		if err := bootstrap.RemoveLinesContaining(commonConfig, "WEB_REAL_IP_HEADER"); err != nil {
			return err
		}
	}

	return os.MkdirAll(env.ValueOrDefaultNonEmpty("APACHE_RUN_DIR", "/tmp/apache2"), 0o755)
}
