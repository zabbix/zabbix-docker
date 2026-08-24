package config

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strconv"
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

// UpdateIndexedParameter collects the prefix_0, prefix_1, ... variables
// into repeated parameters that reference the original environment variables.
// The variables remain in the environment so the service can expand them
// while reading its configuration.
func UpdateIndexedParameter(env bootstrap.Environment, configPath, param, prefix string) error {
	params, err := collectIndexedParams(env, map[string]string{prefix: param})
	if err != nil {
		return err
	}
	if len(params) == 0 {
		return nil
	}

	values := make([]string, 0, len(params))
	variableNames := make([]string, 0, len(params))
	for _, entry := range params {
		values = append(values, "${"+entry.variable+"}")
		variableNames = append(variableNames, entry.variable)
	}
	if _, err := rewriteParameter(configPath, param, values, false); err != nil {
		return err
	}

	bootstrap.LogDebug(env, "** Configuring %s parameter '%s' from %d indexed environment variables: %s",
		configPath, param, len(variableNames), strings.Join(variableNames, ", "),
	)

	return nil
}

func collectIndexedParams(env bootstrap.Environment, paramByPrefix map[string]string) ([]indexedParam, error) {
	for prefix := range paramByPrefix {
		if _, exists := env[prefix]; exists {
			return nil, fmt.Errorf("%s is not supported; use indexed variables such as %s_0", prefix, prefix)
		}
	}

	var params []indexedParam
	for variable, value := range env {
		param, index, found := parseIndexedVariable(variable, paramByPrefix)
		if !found {
			continue
		}
		if value == "" {
			return nil, fmt.Errorf("%s must not be empty", variable)
		}

		params = append(params, indexedParam{
			index:    index,
			param:    param,
			variable: variable,
		})
	}

	sort.Slice(params, func(i, j int) bool {
		if params[i].index == params[j].index {
			return params[i].variable < params[j].variable
		}
		return params[i].index < params[j].index
	})

	for expectedIndex, entry := range params {
		if expectedIndex > 0 && params[expectedIndex-1].index == entry.index {
			return nil, fmt.Errorf(
				"index %d is used by both %s and %s",
				entry.index, params[expectedIndex-1].variable, entry.variable,
			)
		}
		if entry.index != expectedIndex {
			return nil, fmt.Errorf(
				"%s uses index %d, but index %d is missing",
				entry.variable, entry.index, expectedIndex,
			)
		}
	}

	return params, nil
}

func parseIndexedVariable(variable string, paramByPrefix map[string]string) (string, int, bool) {
	separatorIndex := strings.LastIndexByte(variable, '_')
	if separatorIndex == -1 {
		return "", 0, false
	}

	param, found := paramByPrefix[variable[:separatorIndex]]
	if !found {
		return "", 0, false
	}

	indexText := variable[separatorIndex+1:]
	index, err := strconv.Atoi(indexText)
	if err != nil || index < 0 || strconv.Itoa(index) != indexText {
		return "", 0, false
	}

	return param, index, true
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

	_, err = writeLines(configPath, original, updatedLines)
	return err
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
	if isMaskedParameter(param) {
		loggedValue = "****"
	}

	if changed {
		bootstrap.LogInfo("** Updating %s parameter '%s': '%s'", configPath, param, loggedValue)
	} else {
		bootstrap.LogInfo("** Updating %s parameter '%s': '%s'... exists", configPath, param, loggedValue)
	}
}

func isMaskedParameter(param string) bool {
	switch param {
	case "TLSPSKIdentity", "DBPassword", "HistoryProvider":
		return true
	default:
		return false
	}
}
