package bootstrap

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// UpdateConfigValues adds one name option per comma-separated item of rawValue,
// preserving active values already present. An empty value removes the option.
func UpdateConfigValues(configPath, name, rawValue string) error {
	value := strings.Trim(strings.TrimSpace(rawValue), `"`)
	if value == "" {
		changed, err := rewriteConfig(configPath, name, nil, false)
		if err != nil {
			return err
		}
		logConfigChange(configPath, name, nil, changed)

		return nil
	}

	items := strings.Split(value, ",")
	values := make([]string, 0, len(items))
	for _, item := range items {
		if item != "" {
			values = append(values, item)
		}
	}

	changed, err := rewriteConfig(configPath, name, values, true)
	if err != nil {
		return err
	}
	logConfigChange(configPath, name, values, changed)

	return nil
}

// UpdateConfigValue sets a single-value option in the configuration file.
func UpdateConfigValue(configPath, name, value string) error {
	values := []string{value}

	changed, err := rewriteConfig(configPath, name, values, false)
	if err != nil {
		return err
	}
	logConfigChange(configPath, name, values, changed)

	return nil
}

// UpdateConfigIndexed collects the prefix_0, prefix_1, ... variables into
// repeated name options that reference the original environment variables.
// The variables remain in the environment so the service can expand them
// while reading its configuration.
func UpdateConfigIndexed(env Environment, configPath, name, prefix string) error {
	var variables []string
	var values []string

	for index := 0; ; index++ {
		variable := fmt.Sprintf("%s_%d", prefix, index)
		if env[variable] == "" {
			break
		}
		variables = append(variables, variable)
		values = append(values, "${"+variable+"}")
	}

	if len(values) == 0 {
		return nil
	}

	if _, err := rewriteConfig(configPath, name, values, false); err != nil {
		return err
	}

	LogDebug(env, "** Configuring %s parameter '%s' from %d indexed environment variables: %s",
		configPath, name, len(variables), strings.Join(variables, ", "),
	)

	return nil
}

// rewriteConfig writes the name option to the configuration file and reports
// whether the file changed. preserveExisting keeps the option lines that are
// already in the file instead of replacing them.
func rewriteConfig(configPath, name string, values []string, preserveExisting bool) (bool, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false, fmt.Errorf("missing configuration file %s: %w", configPath, err)
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
		return false, nil
	}

	if err := WriteFilePreservingMode(configPath, updatedData); err != nil {
		return false, fmt.Errorf("update configuration file %s: %w", configPath, err)
	}

	return true, nil
}

// logConfigChange reports what happened to the option: values without a
// single non-empty item mean the option was removed from the file.
func logConfigChange(configPath, name string, values []string, changed bool) {
	written := false
	for _, value := range values {
		if value != "" {
			written = true
			break
		}
	}

	if !written {
		if changed {
			LogInfo("** Removing %s parameter '%s'", configPath, name)
		}

		return
	}

	loggedValue := strings.Join(values, ",")
	if isMaskedConfigVar(name) {
		loggedValue = "****"
	}

	if changed {
		LogInfo("** Updating %s parameter '%s': '%s'", configPath, name, loggedValue)
	} else {
		LogInfo("** Updating %s parameter '%s': '%s'... exists", configPath, name, loggedValue)
	}
}

func isMaskedConfigVar(name string) bool {
	switch name {
	case "TLSPSKIdentity", "DBPassword", "HistoryProvider":
		return true
	default:
		return false
	}
}
