package web

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func preparePHP(env bootstrap.Environment, databaseType DatabaseType) error {
	bootstrap.LogInfo("** Preparing PHP configuration")

	homeDir, err := bootstrap.RequiredHomeDirectory(env)
	if err != nil {
		return err
	}

	env.SetDefaultNonEmpty("PHP_FPM_PM", "dynamic")
	env.SetDefaultNonEmpty("PHP_FPM_PM_MAX_CHILDREN", "50")
	env.SetDefaultNonEmpty("PHP_FPM_PM_START_SERVERS", "5")
	env.SetDefaultNonEmpty("PHP_FPM_PM_MIN_SPARE_SERVERS", "5")
	env.SetDefaultNonEmpty("PHP_FPM_PM_MAX_SPARE_SERVERS", "35")
	env.SetDefaultNonEmpty("PHP_FPM_PM_MAX_REQUESTS", "0")

	env.SetDefaultNonEmpty("ZBX_DENY_GUI_ACCESS", "false")
	env.SetDefaultNonEmpty("ZBX_GUI_ACCESS_IP_RANGE", "['127.0.0.1']")
	env.SetDefaultNonEmpty("ZBX_GUI_WARNING_MSG", "Zabbix is under maintenance.")

	env.SetDefaultNonEmpty("ZBX_MAXEXECUTIONTIME", "600")
	env.SetDefaultNonEmpty("ZBX_MEMORYLIMIT", "128M")
	env.SetDefaultNonEmpty("ZBX_POSTMAXSIZE", "16M")
	env.SetDefaultNonEmpty("ZBX_UPLOADMAXFILESIZE", "2M")
	env.SetDefaultNonEmpty("ZBX_MAXINPUTTIME", "300")
	env.SetDefaultNonEmpty("PHP_TZ", "Europe/Riga")

	env.SetDefaultNonEmpty("ZBX_DB_ENCRYPTION", "false")
	env.SetDefaultNonEmpty("ZBX_DB_VERIFY_HOST", "false")

	env.SetDefaultNonEmpty("DB_DOUBLE_IEEE754", "true")

	env.SetDefaultNonEmpty("ZBX_HISTORYPROVIDERS", "[]")

	env.SetDefaultNonEmpty("ZBX_CERT_STORAGE", "database")

	env.SetDefaultNonEmpty("ZBX_BANNERS_ENABLED", "true")
	env.SetDefaultNonEmpty("ZBX_HTTP_AUTH_ENABLED", "true")
	env.SetDefaultNonEmpty("ZBX_MODULES_CONFIG_ENABLED", "true")
	env.SetDefaultNonEmpty("ZBX_MEDIA_TYPE_DENYLIST", "[]")

	env.SetDefaultNonEmpty("ZBX_SERVER_TLS_ACTIVE", "0")

	env.SetDefault("ZBX_SERVER_HOST", "zabbix-server")
	env.SetDefault("ZBX_SERVER_PORT", "10051")

	expose := strings.ToLower(env.ValueOrDefaultNonEmpty("EXPOSE_WEB_SERVER_INFO", "on"))
	if expose != "off" {
		expose = "on"
	}
	env["EXPOSE_WEB_SERVER_INFO"] = expose

	for _, name := range []string{"ZBX_DENY_GUI_ACCESS", "ZBX_DB_ENCRYPTION", "ZBX_DB_VERIFY_HOST", "DB_DOUBLE_IEEE754"} {
		env[name] = strings.ToLower(env[name])
	}

	if os.Getuid() == 0 {
		configFile := env["PHP_ZBX_CONFIG_FILE"]
		if configFile == "" {
			return fmt.Errorf("PHP_ZBX_CONFIG_FILE must be set")
		}

		daemonUser := env["DAEMON_USER"]
		daemonGroup := env["DAEMON_GROUP"]
		userConfig := fmt.Sprintf("user = %s\ngroup = %s\nlisten.owner = %s\nlisten.group = %s\n", daemonUser, daemonGroup, daemonUser, daemonGroup)

		data, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("missing configuration file %s: %w", configFile, err)
		}

		if !strings.Contains(string(data), userConfig) {
			if err := bootstrap.WriteFilePreservingMode(configFile, append(data, userConfig...)); err != nil {
				return err
			}
		}
	}

	env["DB_SERVER_TYPE"] = string(databaseType)
	env["DB_SERVER_USER"] = env["DB_SERVER_ZBX_USER"]
	env["DB_SERVER_PASS"] = env["DB_SERVER_ZBX_PASS"]

	// PHP-FPM expands this variable while parsing the MySQL pool configuration
	if databaseType == DatabaseMySQL {
		env.SetDefault("DB_SERVER_SOCKET", "")
	}

	if err := bootstrap.ProcessTLSFiles(env, homeDir, "ZBX_SERVER_TLS_CA", "ZBX_SERVER_TLS_KEY", "ZBX_SERVER_TLS_CERT"); err != nil {
		return err
	}

	return nil
}

func prepareSessionName(env bootstrap.Environment) error {
	sessionName := env["ZBX_SESSION_NAME"]
	if sessionName == "" {
		return nil
	}

	path := filepath.Join(env["ZABBIX_WWW_ROOT"], "include", "defines.inc.php")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("missing file %s: %w", path, err)
	}

	pattern := regexp.MustCompile(`(ZBX_SESSION_NAME',[[:space:]]*')[^']*('.*)`)
	updated := pattern.ReplaceAllString(string(data), "${1}"+strings.ReplaceAll(sessionName, "$", "$$")+"${2}")

	return bootstrap.WriteFilePreservingMode(path, []byte(updated))
}
