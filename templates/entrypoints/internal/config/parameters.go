package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

type indexedParam struct {
	index    int
	param    string
	variable string
}

// MergeParameterValues adds one parameter per comma-separated item of
// rawValue, preserving active values already present. An empty value removes
// the parameter.
func MergeParameterValues(configPath, param, rawValue string) error {
	value := strings.Trim(strings.TrimSpace(rawValue), `"`)
	if value == "" {
		if err := rewriteParameter(configPath, param, nil, false); err != nil {
			return err
		}

		return nil
	}

	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			values = append(values, part)
		}
	}

	if err := rewriteParameter(configPath, param, values, true); err != nil {
		return err
	}

	return nil
}

// SetParameter sets a single-value parameter in the configuration file.
func SetParameter(configPath, param, value string) error {
	values := []string{value}

	if err := rewriteParameter(configPath, param, values, false); err != nil {
		return err
	}

	return nil
}

// UpdateIndexedParameter collects the prefix_0, prefix_1, ... variables
// into repeated parameters that reference the original environment variables.
// The variables remain in the environment so the service can expand them
// while reading its configuration.
func UpdateIndexedParameter(env bootstrap.Environment, configPath, param, prefix string) error {
	params, err := collectIndexedParams(env, map[string]string{prefix: param}, false)
	if err != nil {
		return err
	}
	if len(params) == 0 {
		return nil
	}

	values := make([]string, 0, len(params))
	for _, entry := range params {
		values = append(values, "${"+entry.variable+"}")
	}
	if err := rewriteParameter(configPath, param, values, false); err != nil {
		return err
	}

	return nil
}

func collectIndexedParams(env bootstrap.Environment, paramByPrefix map[string]string, strict bool) ([]indexedParam, error) {
	prefixes := make([]string, 0, len(paramByPrefix))
	for prefix := range paramByPrefix {
		prefixes = append(prefixes, prefix)
	}

	variables, err := bootstrap.CollectIndexed(env, prefixes, strict)
	if err != nil {
		return nil, err
	}

	params := make([]indexedParam, 0, len(variables))
	for _, variable := range variables {
		params = append(params, indexedParam{
			index:    variable.Index,
			param:    paramByPrefix[variable.Prefix],
			variable: variable.Name,
		})
	}
	return params, nil
}

func replaceIndexedParamsAtEnd(configPath string, paramByPrefix map[string]string, params []indexedParam) error {
	original, lines, err := readLines(configPath)
	if err != nil {
		return err
	}

	paramsToReplace := make(map[string]struct{}, len(paramByPrefix))
	for _, param := range paramByPrefix {
		paramsToReplace[param] = struct{}{}
	}

	updatedLines := make([]string, 0, len(lines)+len(params)+1)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		separatorIndex := strings.IndexByte(trimmed, '=')
		if separatorIndex >= 0 {
			if _, found := paramsToReplace[trimmed[:separatorIndex]]; found {
				continue
			}
		}

		updatedLines = append(updatedLines, line)
	}

	for len(updatedLines) > 0 && strings.TrimSpace(updatedLines[len(updatedLines)-1]) == "" {
		updatedLines = updatedLines[:len(updatedLines)-1]
	}

	if len(params) > 0 {
		if len(updatedLines) > 0 {
			updatedLines = append(updatedLines, "")
		}
		for _, entry := range params {
			updatedLines = append(updatedLines, entry.param+"=${"+entry.variable+"}")
		}
	}

	return writeLines(configPath, original, updatedLines)
}

// rewriteParameter writes the named parameter to the configuration file.
// preserveExisting keeps parameter lines that are already in the file instead
// of replacing them.
func rewriteParameter(configPath, param string, values []string, preserveExisting bool) error {
	original, lines, err := readLines(configPath)
	if err != nil {
		return err
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

func writeLines(configPath string, original []byte, lines []string) error {
	updatedData := []byte(strings.Join(lines, "\n") + "\n")
	if bytes.Equal(original, updatedData) {
		return nil
	}

	if err := bootstrap.WriteFilePreservingMode(configPath, updatedData); err != nil {
		return fmt.Errorf("update configuration file %s: %w", configPath, err)
	}

	return nil
}
