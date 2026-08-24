package web

import (
	"fmt"
	"os"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// phpDefaults are the frontend settings that get a default value when the
// variable is missing or empty.
var phpDefaults = []struct{ name, value string }{
	// PHP-FPM process manager
	{"PHP_FPM_PM", "dynamic"},
	{"PHP_FPM_PM_MAX_CHILDREN", "50"},
	{"PHP_FPM_PM_START_SERVERS", "5"},
	{"PHP_FPM_PM_MIN_SPARE_SERVERS", "5"},
	{"PHP_FPM_PM_MAX_SPARE_SERVERS", "35"},
	{"PHP_FPM_PM_MAX_REQUESTS", "0"},

	// Maintenance mode
	{"ZBX_DENY_GUI_ACCESS", "false"},
	{"ZBX_GUI_ACCESS_IP_RANGE", "['127.0.0.1']"},
	{"ZBX_GUI_WARNING_MSG", "Zabbix is under maintenance."},

	// PHP resource limits
	{"ZBX_MAXEXECUTIONTIME", "600"},
	{"ZBX_MEMORYLIMIT", "128M"},
	{"ZBX_POSTMAXSIZE", "16M"},
	{"ZBX_UPLOADMAXFILESIZE", "2M"},
	{"ZBX_MAXINPUTTIME", "300"},
	{"PHP_TZ", "Europe/Riga"},

	// Database connection
	{"ZBX_DB_ENCRYPTION", "false"},
	{"ZBX_DB_VERIFY_HOST", "false"},
	{"ZBX_DB_DOUBLE_IEEE754", "true"},

	// Frontend features
	{"ZBX_HISTORYPROVIDERS", "[]"},
	{"ZBX_CERT_STORAGE", "database"},
	{"ZBX_BANNERS_ENABLED", "true"},
	{"ZBX_HTTP_AUTH_ENABLED", "true"},
	{"ZBX_MODULES_CONFIG_ENABLED", "true"},
	{"ZBX_MEDIA_TYPE_DENYLIST", "[]"},

	// Connection to Zabbix server
	{"ZBX_SERVER_TLS_ACTIVE", "0"},
}

func preparePHP(env bootstrap.Environment, dbType DBType) error {
	bootstrap.LogInfo("** Preparing PHP configuration")

	homeDir, err := bootstrap.RequiredHomeDir(env)
	if err != nil {
		return err
	}

	for _, setting := range phpDefaults {
		env.SetDefaultNonEmpty(setting.name, setting.value)
	}

	env.SetDefault("ZBX_SERVER_HOST", "zabbix-server")
	env.SetDefault("ZBX_SERVER_PORT", "10051")

	expose := strings.ToLower(env.ValueOrDefaultNonEmpty("EXPOSE_WEB_SERVER_INFO", "on"))
	if expose != "off" {
		expose = "on"
	}
	env["EXPOSE_WEB_SERVER_INFO"] = expose

	for _, name := range []string{"ZBX_DENY_GUI_ACCESS", "ZBX_DB_ENCRYPTION", "ZBX_DB_VERIFY_HOST", "ZBX_DB_DOUBLE_IEEE754"} {
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

	env["DB_SERVER_TYPE"] = string(dbType)

	// PHP-FPM expands this variable while parsing the MySQL pool configuration
	if dbType == MySQL {
		env.SetDefault("DB_SERVER_SOCKET", "")
	}

	if err := bootstrap.ProcessTLSFiles(env, homeDir, "ZBX_SERVER_TLS_CA", "ZBX_SERVER_TLS_KEY", "ZBX_SERVER_TLS_CERT"); err != nil {
		return err
	}

	return nil
}
