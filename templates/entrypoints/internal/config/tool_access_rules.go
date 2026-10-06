package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

var toolAccessRuleParamByPrefix = map[string]string{
	"ZBX_ALLOWTOOL":        "AllowTool",
	"ZBX_DENYTOOL":         "DenyTool",
	"ZBX_ALLOWTOOL_REGEXP": "AllowToolRegexp",
	"ZBX_DENYTOOL_REGEXP":  "DenyToolRegexp",
}

// ConfigureToolAccessRules restores the runtime file from the read-only template
// and replaces its tool access rules only when indexed environment rules are set.
func ConfigureToolAccessRules(env bootstrap.Environment, templatePath, runtimePath string) error {
	rules, err := collectIndexedParams(env, toolAccessRuleParamByPrefix, true)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read tool access template %s: %w", templatePath, err)
	}
	if err := os.WriteFile(runtimePath, data, 0o644); err != nil {
		return fmt.Errorf("write runtime tool access configuration %s: %w", runtimePath, err)
	}
	if len(rules) == 0 {
		return nil
	}
	if err := replaceIndexedParamsAtEnd(runtimePath, toolAccessRuleParamByPrefix, rules); err != nil {
		return err
	}

	variableNames := make([]string, 0, len(rules))
	for _, rule := range rules {
		variableNames = append(variableNames, rule.variable)
	}

	bootstrap.LogDebug(env, "** Configuring %s tool access rules from %d indexed environment variables: %s",
		runtimePath, len(variableNames), strings.Join(variableNames, ", "),
	)

	return nil
}
