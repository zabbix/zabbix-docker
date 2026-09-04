//go:build windows

package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// MergeParameterValues adds one parameter per comma-separated item of
// rawValue, preserving active values already present. An empty value removes
// the parameter.
func MergeParameterValues(configPath, param, rawValue string) error {
	value := strings.Trim(strings.TrimSpace(rawValue), `"`)
	if value == "" {
		changed, err := rewriteParameter(configPath, param, nil, false)
		if err != nil {
			return err
		}
		logParameterChange(configPath, param, nil, changed)

		return nil
	}

	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			values = append(values, part)
		}
	}

	changed, err := rewriteParameter(configPath, param, values, true)
	if err != nil {
		return err
	}
	logParameterChange(configPath, param, values, changed)

	return nil
}

// SetParameter sets a single-value parameter in the configuration file.
func SetParameter(configPath, param, value string) error {
	values := []string{value}

	changed, err := rewriteParameter(configPath, param, values, false)
	if err != nil {
		return err
	}
	logParameterChange(configPath, param, values, changed)

	return nil
}

// rewriteParameter writes the named parameter to the configuration file and
// reports whether the file changed. preserveExisting keeps parameter lines
// that are already in the file instead of replacing them.
func rewriteParameter(configPath, param string, values []string, preserveExisting bool) (bool, error) {
	original, lines, err := readLines(configPath)
	if err != nil {
		return false, err
	}

	activePrefix := param + "="
	commentPrefixes := []string{"# " + activePrefix, "; " + activePrefix}

	updatedLines := make([]string, 0, len(lines)+len(values)+1)
	existingValues := make(map[string]struct{})
	insertAt := -1
	for _, line := range lines {
		if strings.HasPrefix(line, activePrefix) {
			if preserveExisting {
				updatedLines = append(updatedLines, line)
				existingValues[strings.TrimPrefix(line, activePrefix)] = struct{}{}
				insertAt = len(updatedLines)
			}
			continue
		}

		updatedLines = append(updatedLines, line)

		if insertAt == -1 {
			for _, prefix := range commentPrefixes {
				if strings.HasPrefix(line, prefix) {
					insertAt = len(updatedLines)
					break
				}
			}
		}
	}

	paramLines := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}

		if _, found := existingValues[value]; !found {
			paramLines = append(paramLines, activePrefix+value)
		}
	}

	if insertAt >= 0 {
		linesWithParams := make([]string, 0, len(updatedLines)+len(paramLines))
		linesWithParams = append(linesWithParams, updatedLines[:insertAt]...)
		linesWithParams = append(linesWithParams, paramLines...)
		linesWithParams = append(linesWithParams, updatedLines[insertAt:]...)

		updatedLines = linesWithParams
	} else if len(paramLines) > 0 {
		if len(updatedLines) > 0 && updatedLines[len(updatedLines)-1] != "" {
			updatedLines = append(updatedLines, "")
		}

		updatedLines = append(updatedLines, paramLines...)
	}

	return writeLines(configPath, original, updatedLines)
}

func readLines(configPath string) ([]byte, []string, error) {
	original, err := os.ReadFile(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("missing configuration file %s: %w", configPath, err)
	}

	content := strings.ReplaceAll(string(original), "\r\n", "\n")
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")

	return original, lines, nil
}

func writeLines(configPath string, original []byte, lines []string) (bool, error) {
	updatedData := []byte(strings.Join(lines, "\n") + "\n")
	if bytes.Equal(original, updatedData) {
		return false, nil
	}

	if err := bootstrap.WriteFilePreservingMode(configPath, updatedData); err != nil {
		return false, fmt.Errorf("update configuration file %s: %w", configPath, err)
	}

	return true, nil
}

// logParameterChange reports what happened to the parameter: values
// without a single non-empty item mean it was removed from the file.
func logParameterChange(configPath, param string, values []string, changed bool) {
	hasValue := false
	for _, value := range values {
		if value != "" {
			hasValue = true
			break
		}
	}

	if !hasValue {
		if changed {
			bootstrap.LogInfo("** Removing %s parameter '%s'", configPath, param)
		}

		return
	}

	loggedValue := strings.Join(values, ",")
	if changed {
		bootstrap.LogInfo("** Updating %s parameter '%s': '%s'", configPath, param, loggedValue)
	} else {
		bootstrap.LogInfo("** Updating %s parameter '%s': '%s'... exists", configPath, param, loggedValue)
	}
}
