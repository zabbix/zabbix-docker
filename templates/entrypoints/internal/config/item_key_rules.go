package config

import (
	"path/filepath"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

var itemKeyRuleParamByPrefix = map[string]string{
	"ZBX_ALLOWKEY":        "AllowKey",
	"ZBX_DENYKEY":         "DenyKey",
	"ZBX_ALLOWKEY_REGEXP": "AllowKeyRegexp",
	"ZBX_DENYKEY_REGEXP":  "DenyKeyRegexp",
}

// ConfigureItemKeyRules writes the globally ordered indexed AllowKey, DenyKey,
// AllowKeyRegexp and DenyKeyRegexp variables to the end of the item key
// configuration file.
func ConfigureItemKeyRules(env bootstrap.Environment, configDir, fileName string) error {
	rules, err := collectIndexedParams(env, itemKeyRuleParamByPrefix)
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, fileName)
	if err := replaceIndexedParamsAtEnd(configPath, itemKeyRuleParamByPrefix, rules); err != nil {
		return err
	}

	variableNames := make([]string, 0, len(rules))
	for _, rule := range rules {
		variableNames = append(variableNames, rule.variable)
	}
	bootstrap.LogDebug(env, "** Configuring %s item key rules from %d indexed environment variables: %s",
		configPath, len(variableNames), strings.Join(variableNames, ", "),
	)

	return nil
}
