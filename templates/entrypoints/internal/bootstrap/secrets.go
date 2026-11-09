package bootstrap

import (
	"path/filepath"
	"runtime"
)

// SecretComponent identifies an image family in the explicit secret registry.
type SecretComponent string

const (
	SecretsAgent       SecretComponent = "agent"
	SecretsAgent2      SecretComponent = "agent2"
	SecretsServerMySQL SecretComponent = "server-mysql"
	SecretsServerPgSQL SecretComponent = "server-pgsql"
	SecretsProxyMySQL  SecretComponent = "proxy-mysql"
	SecretsProxySQLite SecretComponent = "proxy-sqlite3"
	SecretsWebMySQL    SecretComponent = "web-mysql"
	SecretsWebPgSQL    SecretComponent = "web-pgsql"
	SecretsWebService  SecretComponent = "web-service"
	SecretsMCPServer   SecretComponent = "mcp-server"
	SecretsBridge      SecretComponent = "bridge-adapter"
)

// SecretCommand identifies the entrypoint operation whose secrets are needed.
type SecretCommand string

const (
	SecretsService       SecretCommand = "service"
	SecretsInitDB        SecretCommand = "init_db_only"
	SecretsInitHistory   SecretCommand = "init_db_history_providers"
	SecretsInitTelemetry SecretCommand = "init_db_telemetry"
)

var (
	externalSecretsRoot = platformSecretRoot(false)
	internalSecretsRoot = platformSecretRoot(true)
	dockerSecretsRoot   = platformDockerSecretsRoot()
)

func platformDockerSecretsRoot() string {
	if runtime.GOOS == "windows" {
		return platformSecretRoot(false)
	}
	return SecretsDir
}

// WithResolvedSecrets resolves service secrets before component preparation.
// Initialization handlers resolve their own narrower secret plans.
func WithResolvedSecrets(component SecretComponent, binary string, next Entrypoint) Entrypoint {
	return func(env Environment, args []string) error {
		service := len(args) == 0
		if binary != "" {
			_, service = ServiceArgs(args, binary)
		}
		if service {
			if err := ResolveSecrets(env, component, SecretsService); err != nil {
				return err
			}
		}
		return next(env, args)
	}
}

// WithInitDBSecrets resolves database-initialization secrets before next.
func WithInitDBSecrets(component SecretComponent, next func(Environment) error) func(Environment) error {
	if next == nil {
		return nil
	}
	return func(env Environment) error {
		if err := ResolveSecrets(env, component, SecretsInitDB); err != nil {
			return err
		}
		return next(env)
	}
}

// WithHistoryInitSecrets resolves history-provider initialization secrets before next.
func WithHistoryInitSecrets(component SecretComponent, next func(Environment) (int, error)) func(Environment) (int, error) {
	return withStorageInitSecrets(component, SecretsInitHistory, next)
}

// WithTelemetryInitSecrets resolves telemetry initialization secrets before next.
func WithTelemetryInitSecrets(component SecretComponent, next func(Environment) (int, error)) func(Environment) (int, error) {
	return withStorageInitSecrets(component, SecretsInitTelemetry, next)
}

func withStorageInitSecrets(component SecretComponent, command SecretCommand,
	next func(Environment) (int, error)) func(Environment) (int, error) {
	if next == nil {
		return nil
	}
	return func(env Environment) (int, error) {
		if err := ResolveSecrets(env, component, command); err != nil {
			return 0, err
		}
		return next(env)
	}
}

func platformSecretRoot(internal bool) string {
	name := "secrets"
	if internal {
		name = "secrets_internal"
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(`C:\zabbix`, name)
	}
	return filepath.Join("/run/zabbix", name)
}

// ResolveSecrets applies the explicit registry for one component and command.
// Empty environment variables are treated as unset. A zero-length secret file,
// however, is an explicitly supplied empty value and is preserved.
func ResolveSecrets(env Environment, component SecretComponent, command SecretCommand) error {
	files, values, aliases, err := secretPlan(env, component, command)
	if err != nil {
		return err
	}
	if isServerOrProxy(component) || isWeb(component) {
		dropProviderSecretSources(env, "ZBX_HISTORYPROVIDER_", providerLocationSuffixes)
		dropProviderSecretSources(env, "ZBX_TELEMETRYPROVIDER_", providerLocationSuffixes)
	}
	if isProxy(component) {
		dropProviderSecretSources(env, "ZBX_HISTORYPROVIDER_", providerSecretSuffixes)
	}
	if isWeb(component) {
		dropProviderSecretSources(env, "ZBX_HISTORYPROVIDER_", providerFileSecretSuffixes)
	}
	managed := &managedSecrets{}
	defer managed.close()
	if err := cleanupStaleProviderSecrets(managed, component, command); err != nil {
		return err
	}
	if len(files) == 0 && len(values) == 0 && len(aliases) == 0 {
		return nil
	}
	normalizeEmptySecretSources(env, files, values)

	for _, alias := range aliases {
		if err := normalizeSecretAlias(env, alias); err != nil {
			return err
		}
	}

	for _, spec := range files {
		if err := resolveFileSecret(managed, env, spec); err != nil {
			return err
		}
	}
	for _, spec := range values {
		if err := resolveValueSecret(env, spec); err != nil {
			return err
		}
	}

	if isServerOrProxy(component) {
		switch command {
		case SecretsService, SecretsInitHistory, SecretsInitTelemetry:
			if err := resolveSharedTLSDirectories(env); err != nil {
				return err
			}
			if command == SecretsService {
				if err := prepareVaultTLSFiles(managed, env); err != nil {
					return err
				}
			}
			if isServer(component) && (command == SecretsService || command == SecretsInitHistory) {
				if err := prepareHistoryProviderTLSFiles(managed, env, files); err != nil {
					return err
				}
			}
			if err := resolveGlobalCA(managed, env, files); err != nil {
				return err
			}
		case SecretsInitDB:
			if env["ZBX_VAULTDBPATH"] != "" {
				if err := resolveGlobalCA(managed, env, nil); err != nil {
					return err
				}
			}
		}
	} else if isWeb(component) && command == SecretsService && env["ZBX_VAULTDBPATH"] != "" {
		if err := resolveGlobalCA(managed, env, nil); err != nil {
			return err
		}
	}
	if command == SecretsService && isWeb(component) {
		if err := resolveWebAPMCA(managed, env); err != nil {
			return err
		}
	}
	if isServerOrProxy(component) {
		if err := resolveProviderCAs(managed, env, files); err != nil {
			return err
		}
	}

	return nil
}
