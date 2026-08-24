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

type indexedParameter struct {
	index     int
	parameter string
	variable  string
}

// MergeParameterValues adds one parameter per comma-separated item of
// rawValue, preserving active values already present. An empty value removes
// the parameter.
func MergeParameterValues(configPath, parameter, rawValue string) error {
	value := strings.Trim(strings.TrimSpace(rawValue), `"`)
	if value == "" {
		changed, err := rewriteParameter(configPath, parameter, nil, false)
		if err != nil {
			return err
		}
		logParameterChange(configPath, parameter, nil, changed)

		return nil
	}

	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			values = append(values, part)
		}
	}

	changed, err := rewriteParameter(configPath, parameter, values, true)
	if err != nil {
		return err
	}
	logParameterChange(configPath, parameter, values, changed)

	return nil
}

// SetParameter sets a single-value parameter in the configuration file.
func SetParameter(configPath, parameter, value string) error {
	values := []string{value}

	changed, err := rewriteParameter(configPath, parameter, values, false)
	if err != nil {
		return err
	}
	logParameterChange(configPath, parameter, values, changed)

	return nil
}

// UpdateIndexedParameter collects the prefix_0, prefix_1, ... variables
// into repeated parameters that reference the original environment variables.
// The variables remain in the environment so the service can expand them
// while reading its configuration.
func UpdateIndexedParameter(env bootstrap.Environment, configPath, parameter, prefix string) error {
	parameters, err := collectIndexedParameters(env, map[string]string{prefix: parameter})
	if err != nil {
		return err
	}
	if len(parameters) == 0 {
		return nil
	}

	values := make([]string, 0, len(parameters))
	variableNames := make([]string, 0, len(parameters))
	for _, entry := range parameters {
		values = append(values, "${"+entry.variable+"}")
		variableNames = append(variableNames, entry.variable)
	}
	if _, err := rewriteParameter(configPath, parameter, values, false); err != nil {
		return err
	}

	bootstrap.LogDebug(env, "** Configuring %s parameter '%s' from %d indexed environment variables: %s",
		configPath, parameter, len(variableNames), strings.Join(variableNames, ", "),
	)

	return nil
}

func collectIndexedParameters(env bootstrap.Environment, parameterByPrefix map[string]string) ([]indexedParameter, error) {
	for prefix := range parameterByPrefix {
		if _, exists := env[prefix]; exists {
			return nil, fmt.Errorf("%s is not supported; use indexed variables such as %s_0", prefix, prefix)
		}
	}

	var parameters []indexedParameter
	for variable, value := range env {
		parameter, index, found := parseIndexedVariable(variable, parameterByPrefix)
		if !found {
			continue
		}
		if value == "" {
			return nil, fmt.Errorf("%s must not be empty", variable)
		}

		parameters = append(parameters, indexedParameter{
			index:     index,
			parameter: parameter,
			variable:  variable,
		})
	}

	sort.Slice(parameters, func(i, j int) bool {
		if parameters[i].index == parameters[j].index {
			return parameters[i].variable < parameters[j].variable
		}
		return parameters[i].index < parameters[j].index
	})

	for expectedIndex, entry := range parameters {
		if expectedIndex > 0 && parameters[expectedIndex-1].index == entry.index {
			return nil, fmt.Errorf(
				"index %d is used by both %s and %s",
				entry.index, parameters[expectedIndex-1].variable, entry.variable,
			)
		}
		if entry.index != expectedIndex {
			return nil, fmt.Errorf("index %d is missing", expectedIndex)
		}
	}

	return parameters, nil
}

func parseIndexedVariable(variable string, parameterByPrefix map[string]string) (string, int, bool) {
	separatorIndex := strings.LastIndexByte(variable, '_')
	if separatorIndex == -1 {
		return "", 0, false
	}

	parameter, found := parameterByPrefix[variable[:separatorIndex]]
	if !found {
		return "", 0, false
	}

	indexText := variable[separatorIndex+1:]
	index, err := strconv.Atoi(indexText)
	if err != nil || index < 0 || strconv.Itoa(index) != indexText {
		return "", 0, false
	}

	return parameter, index, true
}

func replaceIndexedParametersAtEnd(configPath string, parameterByPrefix map[string]string, parameters []indexedParameter) error {
	original, lines, err := readLines(configPath)
	if err != nil {
		return err
	}

	parametersToReplace := make(map[string]struct{}, len(parameterByPrefix))
	for _, parameter := range parameterByPrefix {
		parametersToReplace[parameter] = struct{}{}
	}

	updatedLines := make([]string, 0, len(lines)+len(parameters)+1)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		separatorIndex := strings.IndexByte(trimmed, '=')
		if separatorIndex >= 0 {
			if _, found := parametersToReplace[trimmed[:separatorIndex]]; found {
				continue
			}
		}

		updatedLines = append(updatedLines, line)
	}

	for len(updatedLines) > 0 && strings.TrimSpace(updatedLines[len(updatedLines)-1]) == "" {
		updatedLines = updatedLines[:len(updatedLines)-1]
	}

	if len(parameters) > 0 {
		if len(updatedLines) > 0 {
			updatedLines = append(updatedLines, "")
		}
		for _, entry := range parameters {
			updatedLines = append(updatedLines, entry.parameter+"=${"+entry.variable+"}")
		}
	}

	_, err = writeLines(configPath, original, updatedLines)
	return err
}

// rewriteParameter writes the named parameter to the configuration file and
// reports whether the file changed. preserveExisting keeps parameter lines
// that are already in the file instead of replacing them.
func rewriteParameter(configPath, parameter string, values []string, preserveExisting bool) (bool, error) {
	original, lines, err := readLines(configPath)
	if err != nil {
		return false, err
	}

	activePrefix := parameter + "="
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

	parameterLines := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}

		if _, found := existingValues[value]; !found {
			parameterLines = append(parameterLines, activePrefix+value)
		}
	}

	if insertAt >= 0 {
		linesWithParameters := make([]string, 0, len(updatedLines)+len(parameterLines))
		linesWithParameters = append(linesWithParameters, updatedLines[:insertAt]...)
		linesWithParameters = append(linesWithParameters, parameterLines...)
		linesWithParameters = append(linesWithParameters, updatedLines[insertAt:]...)

		updatedLines = linesWithParameters
	} else if len(parameterLines) > 0 {
		if len(updatedLines) > 0 && updatedLines[len(updatedLines)-1] != "" {
			updatedLines = append(updatedLines, "")
		}

		updatedLines = append(updatedLines, parameterLines...)
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
func logParameterChange(configPath, parameter string, values []string, changed bool) {
	hasValue := false
	for _, value := range values {
		if value != "" {
			hasValue = true
			break
		}
	}

	if !hasValue {
		if changed {
			bootstrap.LogInfo("** Removing %s parameter '%s'", configPath, parameter)
		}

		return
	}

	loggedValue := strings.Join(values, ",")
	if isMaskedParameter(parameter) {
		loggedValue = "****"
	}

	if changed {
		bootstrap.LogInfo("** Updating %s parameter '%s': '%s'", configPath, parameter, loggedValue)
	} else {
		bootstrap.LogInfo("** Updating %s parameter '%s': '%s'... exists", configPath, parameter, loggedValue)
	}
}

func isMaskedParameter(parameter string) bool {
	switch parameter {
	case "TLSPSKIdentity", "DBPassword", "HistoryProvider":
		return true
	default:
		return false
	}
}
