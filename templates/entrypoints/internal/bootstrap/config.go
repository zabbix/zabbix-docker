//go:build windows

package bootstrap

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// ConfigValue describes one Zabbix configuration parameter update.
type ConfigValue struct {
	Name  string
	Value string
}

// UpdateConfigMultiple replaces the name option in the configuration file
// with one line per comma-separated item of rawValue. An empty value removes
// the option.
func UpdateConfigMultiple(configPath, name, rawValue string) error {
	value := strings.Trim(strings.TrimSpace(rawValue), `"`)
	if value == "" {
		return rewriteConfig(configPath, name, nil, false)
	}

	items := strings.Split(value, ",")
	values := make([]string, 0, len(items))
	for _, item := range items {
		if item != "" {
			values = append(values, item)
		}
	}
	return rewriteConfig(configPath, name, values, true)
}

// UpdateConfigValue sets a single-value option in the configuration file.
func UpdateConfigValue(configPath, name, value string) error {
	if value == `""` {
		return rewriteConfig(configPath, name, []string{""}, false)
	}
	if strings.TrimSpace(value) == "" {
		return rewriteConfig(configPath, name, nil, false)
	}

	return rewriteConfig(configPath, name, []string{value}, false)
}

// UpdateConfigValues applies configuration parameter updates in order.
func UpdateConfigValues(configPath string, values ...ConfigValue) error {
	for _, value := range values {
		if err := UpdateConfigValue(configPath, value.Name, value.Value); err != nil {
			return err
		}
	}

	return nil
}

func rewriteConfig(configPath, name string, values []string, preserveExisting bool) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("missing configuration file %s: %w", configPath, err)
	}

	lineEnding := "\n"
	if bytes.Contains(data, []byte("\r\n")) {
		lineEnding = "\r\n"
	}
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")

	activePrefix := name + "="
	commentPrefixes := []string{"# " + activePrefix, "; " + activePrefix}

	output := make([]string, 0, len(lines)+len(values)+1)
	existing := make(map[string]struct{})
	insertAt := -1
	for _, line := range lines {
		if strings.HasPrefix(line, activePrefix) {
			if preserveExisting {
				output = append(output, line)
				existing[strings.TrimPrefix(line, activePrefix)] = struct{}{}
				insertAt = len(output)
			}
			continue
		}

		output = append(output, line)

		if insertAt == -1 && hasAnyPrefix(line, commentPrefixes) {
			insertAt = len(output)
		}
	}

	allowEmpty := len(values) == 1 && values[0] == ""
	newLines := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" && !allowEmpty {
			continue
		}

		if _, found := existing[value]; !found {
			newLines = append(newLines, activePrefix+value)
		}
	}

	if insertAt >= 0 {
		updated := make([]string, 0, len(output)+len(newLines))
		updated = append(updated, output[:insertAt]...)
		updated = append(updated, newLines...)
		updated = append(updated, output[insertAt:]...)

		output = updated
	} else if len(newLines) > 0 {
		if len(output) > 0 && output[len(output)-1] != "" {
			output = append(output, "")
		}

		output = append(output, newLines...)
	}

	requested := len(values) > 0

	updatedData := []byte(strings.ReplaceAll(strings.Join(output, "\n")+"\n", "\n", lineEnding))

	changed := !bytes.Equal(data, updatedData)
	if changed {
		if err := WriteFilePreservingMode(configPath, updatedData); err != nil {
			return fmt.Errorf("update configuration file %s: %w", configPath, err)
		}
	}

	if !requested {
		if changed {
			LogInfo("** Removing %s parameter '%s'", configPath, name)
		}
		return nil
	}

	loggedValue := strings.Join(values, ",")

	if name == "TLSPSKIdentity" {
		loggedValue = "****"
	}

	if changed {
		LogInfo("** Updating %s parameter '%s': '%s'", configPath, name, loggedValue)
	} else {
		LogInfo("** Updating %s parameter '%s': '%s'... exists", configPath, name, loggedValue)
	}

	return nil
}
