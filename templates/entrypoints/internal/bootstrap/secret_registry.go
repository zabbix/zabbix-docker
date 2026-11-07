package bootstrap

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type fileSecretSpec struct {
	inline    string
	file      string
	category  string
	legacyDir string
	standard  bool
	active    bool
}

type valueSecretSpec struct {
	name     string
	category string
	active   bool
}

type secretAlias struct {
	legacy    string
	canonical string
}

func secretPlan(env Environment, component SecretComponent, command SecretCommand) ([]fileSecretSpec, []valueSecretSpec, []secretAlias, error) {
	home := env["ZABBIX_USER_HOME_DIR"]
	enc := filepath.Join(home, "enc")
	webCerts := WebCertsDir

	var files []fileSecretSpec
	addFiles := func(category, legacy string, standard, active bool, names ...string) {
		for _, name := range names {
			files = append(files, fileSecretSpec{inline: name, file: name + "FILE", category: category, legacyDir: legacy, standard: standard, active: active})
		}
	}
	var values []valueSecretSpec
	addValues := func(category string, active bool, names ...string) {
		for _, name := range names {
			values = append(values, valueSecretSpec{name: name, category: category, active: active})
		}
	}
	var aliases []secretAlias

	service := command == SecretsService
	server := isServer(component)
	proxy := isProxy(component)
	web := isWeb(component)

	if service {
		switch component {
		case SecretsAgent, SecretsAgent2:
			addFiles("tls", enc, false, true, "ZBX_TLSCA", "ZBX_TLSCRL", "ZBX_TLSCERT", "ZBX_TLSKEY", "ZBX_TLSPSK")
		case SecretsWebService:
			addFiles("tls", enc, false, secretModeEnabled(env, "cert", "ZBX_TLSACCEPT"), "ZBX_TLSCA", "ZBX_TLSCERT", "ZBX_TLSKEY")
		case SecretsMCPServer:
			addFiles("tls", enc, false, secretModeEnabled(env, "cert", "ZBX_TLSACCEPT"), "ZBX_TLSCERT", "ZBX_TLSKEY")
		case SecretsBridge:
			listenerTLS := strings.EqualFold(env["ZBX_TLSMODE"], "tls") || strings.EqualFold(env["ZBX_TLSMODE"], "mtls")
			listenerMTLS := strings.EqualFold(env["ZBX_TLSMODE"], "mtls")
			apiTLS := strings.EqualFold(env["ZBX_ZABBIXAPITLSCONNECT"], "verify_ca") ||
				strings.EqualFold(env["ZBX_ZABBIXAPITLSCONNECT"], "verify_full")
			addFiles("tls", enc, false, listenerMTLS, "ZBX_TLSCA")
			addFiles("tls", enc, false, listenerTLS, "ZBX_TLSCERT", "ZBX_TLSKEY")
			addFiles("tls", enc, false, apiTLS, "ZBX_ZABBIXAPITLSCA")
		}
		if server {
			addFiles("tls", enc, false, true, "ZBX_TLSCA", "ZBX_TLSCRL", "ZBX_TLSCERT", "ZBX_TLSKEY")
		}
		if proxy {
			addFiles("tls", enc, false, true, "ZBX_TLSCA", "ZBX_TLSCRL", "ZBX_TLSCERT", "ZBX_TLSKEY", "ZBX_TLSPSK")
			addFiles("apm", "", false, envEnabled(env["ZBX_STARTAPMCOLLECTOR"]), "ZBX_APMTLSCA", "ZBX_APMTLSCERT", "ZBX_APMTLSKEY")
		}
		if web {
			addFiles("tls", enc, false, envEnabled(env["ZBX_SERVER_TLS_ACTIVE"]), "ZBX_SERVER_TLS_CA", "ZBX_SERVER_TLS_CERT", "ZBX_SERVER_TLS_KEY")
			addFiles("sso", webCerts, false, true, "ZBX_SSO_SP_KEY", "ZBX_SSO_SP_CERT", "ZBX_SSO_IDP_CERT")
			apmActive := env["ZBX_APM_CA"] != "" || env["ZBX_APM_CAFILE"] != "" || env["ZBX_APM_CA_LOCATION"] != ""
			addFiles("apm", "", false, apmActive, "ZBX_APM_CA")
		}
	}

	if service || command == SecretsInitDB {
		if server || component == SecretsProxyMySQL || web {
			dbTLSActive := env["ZBX_DBTLSCONNECT"] != ""
			if web {
				dbTLSActive = envEnabled(env["ZBX_DB_ENCRYPTION"])
			}
			for _, name := range []string{"ZBX_DBTLSCA", "ZBX_DBTLSCERT", "ZBX_DBTLSKEY"} {
				files = append(files, fileSecretSpec{inline: name, file: name + "FILE", category: "database", standard: true, active: dbTLSActive})
			}
		}
		localDBCredentials := env["ZBX_VAULTDBPATH"] == ""
		if component == SecretsServerMySQL || component == SecretsProxyMySQL || component == SecretsWebMySQL {
			addValues("database", localDBCredentials, "MYSQL_USER", "MYSQL_PASSWORD")
			addValues("database", true, "MYSQL_ROOT_USER", "MYSQL_ROOT_PASSWORD")
		}
		if component == SecretsServerPgSQL || component == SecretsWebPgSQL {
			addValues("database", localDBCredentials, "POSTGRES_USER", "POSTGRES_PASSWORD")
		}
	}

	if server || proxy || web {
		if service || command == SecretsInitDB || command == SecretsInitTelemetry {
			legacy := enc
			if web {
				legacy = webCerts
			}
			vaultActive := env["ZBX_VAULT"] != ""
			if command == SecretsInitTelemetry {
				vaultActive = vaultActive || telemetryProviderUsesVault(env)
			}
			for _, name := range []string{"ZBX_VAULTTLSCERT", "ZBX_VAULTTLSKEY"} {
				files = append(files, fileSecretSpec{inline: name, file: name + "FILE", category: "vault", legacyDir: legacy, active: vaultActive})
			}
			for _, name := range []string{"VAULT_TOKEN", "ZBX_VAULTAPPROLEID", "ZBX_VAULTAPPSECRETID"} {
				values = append(values, valueSecretSpec{name: name, category: "vault", active: vaultActive})
			}
			aliases = append(aliases,
				secretAlias{legacy: "ZBX_VAULTCERTFILE", canonical: "ZBX_VAULTTLSCERTFILE"},
				secretAlias{legacy: "ZBX_VAULTKEYFILE", canonical: "ZBX_VAULTTLSKEYFILE"},
			)
		}
	}
	if web && service {
		aliases = append(aliases,
			secretAlias{legacy: "ZBX_DB_CA_FILE", canonical: "ZBX_DBTLSCAFILE"},
			secretAlias{legacy: "ZBX_DB_CERT_FILE", canonical: "ZBX_DBTLSCERTFILE"},
			secretAlias{legacy: "ZBX_DB_KEY_FILE", canonical: "ZBX_DBTLSKEYFILE"},
		)
	}

	if (server && (service || command == SecretsInitHistory || command == SecretsInitTelemetry)) ||
		(proxy && (service || command == SecretsInitTelemetry)) || (web && service) {
		providerFiles, providerValues, err := providerSecretPlan(env, command, server || web, server)
		if err != nil {
			return nil, nil, nil, err
		}
		files = append(files, providerFiles...)
		values = append(values, providerValues...)
	}

	return files, values, aliases, nil
}

func dropProviderSecretSources(env Environment, prefix string, suffixes []string) {
	for name := range env {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		if hasProviderSecretSuffix(name, suffixes) {
			delete(env, name)
		}
	}
}

func hasProviderSecretSuffix(name string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

func secretModeEnabled(env Environment, mode string, names ...string) bool {
	for _, name := range names {
		for value := range strings.SplitSeq(env[name], ",") {
			if strings.EqualFold(strings.TrimSpace(value), mode) {
				return true
			}
		}
	}
	return false
}

func envEnabled(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func providerSecretPlan(env Environment, command SecretCommand, includeHistory, includeHistoryFiles bool) ([]fileSecretSpec, []valueSecretSpec, error) {
	bases := make(map[string]string)
	for name := range env {
		for _, prefix := range []string{"ZBX_HISTORYPROVIDER_", "ZBX_TELEMETRYPROVIDER_"} {
			if !strings.HasPrefix(name, prefix) {
				continue
			}
			if hasProviderSecretSuffix(name, providerLocationSuffixes) {
				continue
			}
			if prefix == "ZBX_HISTORYPROVIDER_" && !includeHistory {
				continue
			}
			if prefix == "ZBX_HISTORYPROVIDER_" && !includeHistoryFiles &&
				hasProviderSecretSuffix(name, providerFileSecretSuffixes) {
				continue
			}
			rest := strings.TrimPrefix(name, prefix)
			index, _, found := strings.Cut(rest, "_")
			if index == "" {
				continue
			}
			if _, valid := parseIndexedNumber(index); !valid {
				return nil, nil, fmt.Errorf("invalid indexed variable %s", name)
			}
			base := prefix + index
			family := "history"
			if prefix == "ZBX_TELEMETRYPROVIDER_" {
				family = "telemetry"
			}
			if found || name == base {
				bases[base] = "providers/" + family + "-" + index
			}
		}
	}
	keys := make([]string, 0, len(bases))
	for base := range bases {
		if command == SecretsInitHistory && !strings.HasPrefix(base, "ZBX_HISTORYPROVIDER_") {
			continue
		}
		if command == SecretsInitTelemetry && !strings.HasPrefix(base, "ZBX_TELEMETRYPROVIDER_") {
			continue
		}
		keys = append(keys, base)
	}
	sort.Strings(keys)

	var files []fileSecretSpec
	var values []valueSecretSpec
	for _, base := range keys {
		category := bases[base]
		historyWithoutFiles := strings.HasPrefix(base, "ZBX_HISTORYPROVIDER_") && !includeHistoryFiles
		if env[base] == "" {
			if source := configuredProviderSecretSource(env, base, !historyWithoutFiles); source != "" {
				return nil, nil, fmt.Errorf("%s requires a matching non-empty provider variable %s", source, base)
			}
		}
		for _, secret := range providerSecretDefinitions {
			if historyWithoutFiles && secret.Kind == ProviderFileSecret {
				continue
			}
			name := base + secret.Suffix
			switch secret.Kind {
			case ProviderFileSecret:
				files = append(files, fileSecretSpec{inline: name, file: name + "FILE", category: category, active: env[base] != ""})
			case ProviderValueSecret:
				values = append(values, valueSecretSpec{name: name, category: category, active: env[base] != ""})
			}
		}
	}
	return files, values, nil
}

func telemetryProviderUsesVault(env Environment) bool {
	const prefix = "ZBX_TELEMETRYPROVIDER_"
	for name, value := range env {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		index := strings.TrimPrefix(name, prefix)
		if _, valid := parseIndexedNumber(index); !valid {
			continue
		}
		decoded, err := DecodeJSON(value)
		if err != nil {
			continue
		}
		fields, ok := decoded.(map[string]any)
		if !ok {
			continue
		}
		vaultPath, ok := fields["vault_path"].(string)
		if ok && vaultPath != "" {
			return true
		}
	}
	return false
}

func isServerOrProxy(component SecretComponent) bool {
	return isServer(component) || isProxy(component)
}

func isServer(component SecretComponent) bool {
	return component == SecretsServerMySQL || component == SecretsServerPgSQL
}

func isProxy(component SecretComponent) bool {
	return component == SecretsProxyMySQL || component == SecretsProxySQLite
}

func isWeb(component SecretComponent) bool {
	return component == SecretsWebMySQL || component == SecretsWebPgSQL
}
