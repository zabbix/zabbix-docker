package bootstrap

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// IndexedVariable is one PREFIX_N environment variable.
type IndexedVariable struct {
	Index  int
	Name   string
	Prefix string
}

// IndexedName returns the environment variable name for prefix and index.
func IndexedName(prefix string, index int) string {
	return prefix + "_" + strconv.Itoa(index)
}

// CollectIndexed returns a consecutive, globally ordered list of PREFIX_N
// variables. Prefixes do not include the underscore before the index.
// Unknown suffixes are ignored unless strict is true.
func CollectIndexed(env Environment, prefixes []string, strict bool) ([]IndexedVariable, error) {
	orderedPrefixes := append([]string(nil), prefixes...)
	sort.Slice(orderedPrefixes, func(i, j int) bool {
		return len(orderedPrefixes[i]) > len(orderedPrefixes[j])
	})

	for _, prefix := range orderedPrefixes {
		if _, exists := env[prefix]; exists {
			return nil, fmt.Errorf("%s is not supported; use indexed variables such as %s", prefix,
				IndexedName(prefix, 0))
		}
	}

	var variables []IndexedVariable
	for name, value := range env {
		prefix, suffix, found := indexedSuffix(name, orderedPrefixes)
		if !found {
			continue
		}

		index, err := strconv.Atoi(suffix)
		if err != nil || index < 0 || strconv.Itoa(index) != suffix {
			if strict {
				return nil, fmt.Errorf("invalid indexed variable %s", name)
			}
			continue
		}
		if value == "" {
			return nil, fmt.Errorf("%s must not be empty", name)
		}

		variables = append(variables, IndexedVariable{Index: index, Name: name, Prefix: prefix})
	}

	sort.Slice(variables, func(i, j int) bool {
		if variables[i].Index == variables[j].Index {
			return variables[i].Name < variables[j].Name
		}
		return variables[i].Index < variables[j].Index
	})

	for expected, variable := range variables {
		if expected > 0 && variables[expected-1].Index == variable.Index {
			return nil, fmt.Errorf("index %d is used by both %s and %s", variable.Index,
				variables[expected-1].Name, variable.Name)
		}
		if variable.Index != expected {
			return nil, fmt.Errorf("%s uses index %d, but index %d is missing", variable.Name,
				variable.Index, expected)
		}
	}

	return variables, nil
}

func indexedSuffix(name string, prefixes []string) (string, string, bool) {
	for _, prefix := range prefixes {
		if suffix, found := strings.CutPrefix(name, prefix+"_"); found {
			return prefix, suffix, true
		}
	}

	return "", "", false
}
