package agent

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

type itemKeyRule struct {
	index    int
	option   string
	variable string
}

var itemKeyRuleOptions = map[string]string{
	"ZBX_ALLOWKEY":        "AllowKey",
	"ZBX_DENYKEY":         "DenyKey",
	"ZBX_ALLOWKEY_REGEXP": "AllowKeyRegexp",
	"ZBX_DENYKEY_REGEXP":  "DenyKeyRegexp",
}

// ConfigureItemKeyRules writes the globally ordered indexed AllowKey, DenyKey,
// AllowKeyRegexp and DenyKeyRegexp variables to the end of the item key
// configuration file.
func ConfigureItemKeyRules(env bootstrap.Environment, configDir, fileName string) error {
	for variable := range itemKeyRuleOptions {
		if env[variable] != "" {
			return fmt.Errorf("%s is not supported; use indexed variables such as %s_0", variable, variable)
		}
		delete(env, variable)
	}

	rules, err := collectItemKeyRules(env)
	if err != nil {
		return err
	}

	path := filepath.Join(configDir, fileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("missing configuration file %s: %w", path, err)
	}

	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	output := make([]string, 0, len(lines)+len(rules)+1)
	for _, line := range lines {
		if !isItemKeyRuleLine(line) {
			output = append(output, line)
		}
	}
	for len(output) > 0 && strings.TrimSpace(output[len(output)-1]) == "" {
		output = output[:len(output)-1]
	}

	if len(rules) > 0 {
		if len(output) > 0 {
			output = append(output, "")
		}
		for _, rule := range rules {
			output = append(output, rule.option+"=${"+rule.variable+"}")
		}
	}

	updatedData := []byte(strings.Join(output, "\n") + "\n")
	if !bytes.Equal(data, updatedData) {
		if err := bootstrap.WriteFilePreservingMode(path, updatedData); err != nil {
			return fmt.Errorf("update configuration file %s: %w", path, err)
		}
	}

	if len(rules) > 0 {
		variables := make([]string, 0, len(rules))
		for _, rule := range rules {
			variables = append(variables, rule.variable)
		}
		bootstrap.LogDebug(env, "** Configuring %s item key rules from %d indexed environment variables: %s",
			path, len(variables), strings.Join(variables, ", "),
		)
	}

	return nil
}

func collectItemKeyRules(env bootstrap.Environment) ([]itemKeyRule, error) {
	var rules []itemKeyRule
	for variable, value := range env {
		option, index, found := itemKeyRuleVariable(variable)

		if !found {
			continue
		}

		if value == "" {
			return nil, fmt.Errorf("%s must not be empty", variable)
		}

		rules = append(rules, itemKeyRule{
			index:    index,
			option:   option,
			variable: variable,
		})
	}

	sort.Slice(rules, func(i, j int) bool {
		if rules[i].index == rules[j].index {
			return rules[i].variable < rules[j].variable
		}
		return rules[i].index < rules[j].index
	})

	for position, rule := range rules {
		if position > 0 && rules[position-1].index == rule.index {
			return nil, fmt.Errorf(
				"item key rule index %d is used by both %s and %s",
				rule.index, rules[position-1].variable, rule.variable,
			)
		}
		if rule.index != position {
			return nil, fmt.Errorf("item key rule index %d is missing", position)
		}
	}

	return rules, nil
}

// itemKeyRuleVariable parses an indexed item key rule environment variable
// and returns its configuration option, zero-based index and whether the name
// has the supported <prefix>_<index> format.
func itemKeyRuleVariable(variable string) (string, int, bool) {
	separator := strings.LastIndexByte(variable, '_')
	if separator == -1 {
		return "", 0, false
	}

	option, found := itemKeyRuleOptions[variable[:separator]]
	if !found {
		return "", 0, false
	}

	suffix := variable[separator+1:]
	index, err := strconv.Atoi(suffix)
	return option, index, err == nil && index >= 0 && strconv.Itoa(index) == suffix
}

func isItemKeyRuleLine(line string) bool {
	line = strings.TrimSpace(line)

	for _, option := range itemKeyRuleOptions {
		if strings.HasPrefix(line, option+"=") {
			return true
		}
	}

	return false
}
