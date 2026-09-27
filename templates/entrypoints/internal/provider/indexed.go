// Package provider handles provider variables shared by the server, proxy and
// web image entrypoints.
package provider

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// secretSuffixes are the credential variable suffixes, longest first so that a
// shorter suffix is never trimmed from the middle of a longer one.
var secretSuffixes = credentialSuffixes()

func credentialSuffixes() []string {
	suffixes := make([]string, 0, len(ClickHouseSecretFields)*2)
	for _, secret := range ClickHouseSecretFields {
		suffixes = append(suffixes, secret.Suffix, secret.Suffix+"_FILE")
	}
	sort.Slice(suffixes, func(i, j int) bool { return len(suffixes[i]) > len(suffixes[j]) })
	return suffixes
}

// IndexedNames returns consecutive provider names and checks that credentials
// are attached to a provider.
func IndexedNames(env bootstrap.Environment, prefix string) ([]string, error) {
	providerEnv := make(bootstrap.Environment)
	for name, value := range env {
		if name != prefix && !strings.HasPrefix(name, prefix+"_") {
			continue
		}
		if base, isCredential := credentialBase(name); isCredential {
			if value == "" {
				continue
			}
			if _, exists := env[base]; !exists {
				return nil, fmt.Errorf("%s credentials require a matching JSON provider", base)
			}
			continue
		}
		providerEnv[name] = value
	}

	variables, err := bootstrap.CollectIndexed(providerEnv, []string{prefix}, true)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(variables))
	for index, variable := range variables {
		names[index] = variable.Name
	}
	return names, nil
}

// credentialBase returns the provider variable a credential variable belongs to.
func credentialBase(name string) (string, bool) {
	for _, suffix := range secretSuffixes {
		if base, found := strings.CutSuffix(name, suffix); found {
			return base, true
		}
	}
	return name, false
}
