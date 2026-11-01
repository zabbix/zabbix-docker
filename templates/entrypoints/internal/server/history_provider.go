package server

import (
	"fmt"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/storageinit"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/storageprovider"
)

func configureJSONHistoryProviders(env bootstrap.Environment) error {
	if strings.TrimSpace(env[storageprovider.History.Plural]) != "" {
		return fmt.Errorf("%s is not supported by the server; use %s",
			storageprovider.History.Plural, bootstrap.IndexedName(storageprovider.History.Prefix, 0))
	}
	return storageinit.ConfigureHistory(env)
}
