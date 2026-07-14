package bootstrap

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

func UpdateConfigMultiple(path, name, rawValue string) error {
	value := strings.Trim(strings.TrimSpace(rawValue), `"`)
	if value == "" {
		return rewriteConfig(path, name, nil, false)
	}

	items := strings.Split(value, ",")
	values := make([]string, 0, len(items))
	for _, item := range items {
		if item != "" {
			values = append(values, item)
		}
	}
	return rewriteConfig(path, name, values, true)
}

func UpdateConfigValue(path, name, value string) error {
	return rewriteConfig(path, name, []string{value}, false)
}

func UpdateConfigIndexed(env Environment, path, name, prefix string) error {
	var values []string
	for index := 0; ; index++ {
		variable := fmt.Sprintf("%s_%d", prefix, index)
		value := env[variable]
		if value == "" {
			break
		}
		values = append(values, value)
		delete(env, variable)
	}
	if len(values) == 0 {
		return nil
	}
	return rewriteConfig(path, name, values, false)
}

func rewriteConfig(path, name string, values []string, preserveExisting bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("missing configuration file %s: %w", path, err)
	}

	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
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

	newLines := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
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

	updatedData := []byte(strings.Join(output, "\n") + "\n")
	if bytes.Equal(data, updatedData) {
		return nil
	}
	if err := os.WriteFile(path, updatedData, 0o644); err != nil {
		return fmt.Errorf("update configuration file %s: %w", path, err)
	}
	if len(newLines) == 0 {
		LogInfo("** Removing %s parameter '%s'", path, name)
		return nil
	}
	loggedValue := strings.Join(values, ",")
	if isMaskedConfigVariable(name) && loggedValue != "" {
		loggedValue = "****"
	}
	LogInfo("** Updating %s parameter '%s': '%s'", path, name, loggedValue)
	return nil
}

func isMaskedConfigVariable(name string) bool {
	switch name {
	case "TLSPSKIdentity", "DBPassword", "HistoryProvider":
		return true
	default:
		return false
	}
}
