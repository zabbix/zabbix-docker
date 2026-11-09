package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolveSharedTLSDirectories(env Environment) error {
	home := env["ZABBIX_USER_HOME_DIR"]
	pairs := []struct {
		envName, category, legacy string
	}{
		{"ZBX_SSLCERTLOCATION", "ssl/certs", filepath.Join(home, "ssl", "certs")},
		{"ZBX_SSLKEYLOCATION", "ssl/keys", filepath.Join(home, "ssl", "keys")},
	}
	for _, pair := range pairs {
		resolved, err := resolveSecretDirectory(pair.envName, filepath.Join(externalSecretsRoot, filepath.FromSlash(pair.category)), pair.legacy)
		if err != nil {
			return err
		}
		if resolved != "" {
			env[pair.envName] = resolved
		}
	}
	return nil
}

type managedCAFile struct {
	name string
	path string
}

func resolveGlobalCA(managed *managedSecrets, env Environment, files []fileSecretSpec) error {
	var extras []managedCAFile
	var historySpecs []fileSecretSpec
	for _, spec := range files {
		if !strings.HasPrefix(spec.inline, "ZBX_HISTORYPROVIDER_") || !strings.HasSuffix(spec.inline, "_SSL_CA") {
			continue
		}
		if path := env[spec.file]; path != "" {
			extras = append(extras, managedCAFile{
				name: generatedProviderCAFileName(spec.file, path),
				path: path,
			})
			historySpecs = append(historySpecs, spec)
		}
	}
	if env["ZBX_SSLCALOCATION"] != "" && len(extras) == 0 {
		return nil
	}

	home := env["ZABBIX_USER_HOME_DIR"]
	source := env["ZBX_SSLCALOCATION"]
	if source == "" {
		var err error
		source, err = resolveSecretDirectory(
			"ZBX_SSLCALOCATION",
			filepath.Join(externalSecretsRoot, "ssl", "ca"),
			filepath.Join(home, "ssl", "ssl_ca"),
		)
		if err != nil {
			return err
		}
	}
	target := filepath.Join(internalSecretsRoot, "ssl", "ca")
	used, err := directoryHasFiles(source)
	if err != nil {
		return err
	}
	if !used && len(extras) == 0 {
		delete(env, "ZBX_SSLCALOCATION")
		return managed.removeTree(target)
	}
	if err := prepareManagedCADir(managed, target, []string{source}, extras); err != nil {
		return err
	}
	for _, spec := range historySpecs {
		delete(env, spec.file)
		delete(env, strings.TrimSuffix(spec.file, "FILE")+"LOCATION")
	}
	env["ZBX_SSLCALOCATION"] = target
	return nil
}

func resolveWebAPMCA(managed *managedSecrets, env Environment) error {
	if env["ZBX_APM_CA_LOCATION"] != "" {
		return nil
	}
	current := filepath.Join(externalSecretsRoot, "apm", "ca")
	currentUsed, err := directoryHasFiles(current)
	if err != nil {
		return err
	}
	if !currentUsed {
		delete(env, "ZBX_APM_CA_LOCATION")
		return managed.removeTree(filepath.Join(internalSecretsRoot, "apm", "ca"))
	}
	target := filepath.Join(internalSecretsRoot, "apm", "ca")
	if err := prepareManagedCADir(managed, target, []string{current}, nil); err != nil {
		return err
	}
	env["ZBX_APM_CA_LOCATION"] = target
	return nil
}

func resolveProviderCAs(managed *managedSecrets, env Environment, files []fileSecretSpec) error {
	for _, spec := range files {
		if !strings.HasPrefix(spec.inline, "ZBX_TELEMETRYPROVIDER_") || !strings.HasSuffix(spec.inline, "_SSL_CA") {
			continue
		}
		path := env[spec.file]
		if path == "" {
			continue
		}
		target := filepath.Join(internalSecretsRoot, filepath.FromSlash(spec.category), "ca")
		extra := []managedCAFile{{name: generatedProviderCAFileName(spec.file, path), path: path}}
		if err := prepareManagedCADir(managed, target, nil, extra); err != nil {
			return fmt.Errorf("%s: %w", spec.file, err)
		}
		env[strings.TrimSuffix(spec.file, "FILE")+"LOCATION"] = target
	}
	return nil
}

func prepareManagedCADir(managed *managedSecrets, target string, sources []string, files []managedCAFile) error {
	type content struct {
		data   []byte
		source string
	}
	contents := make(map[string]content)
	copyOne := func(name, source string) error {
		if source == "" || name == "" {
			return nil
		}
		if previous, exists := contents[name]; exists {
			return fmt.Errorf("CA file name %q is supplied by both %q and %q", name, previous.source, source)
		}
		data, err := os.ReadFile(source)
		if err != nil {
			return fmt.Errorf("read CA file %q: %w", source, err)
		}
		contents[name] = content{data: data, source: source}
		return nil
	}
	for _, sourceDir := range sources {
		if sourceDir == "" {
			continue
		}
		entries, err := os.ReadDir(sourceDir)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read CA source directory %q: %w", sourceDir, err)
		}
		for _, entry := range entries {
			if filepath.Clean(sourceDir) == filepath.Clean(target) {
				if strings.HasPrefix(entry.Name(), "ZBX_HISTORYPROVIDER_") || !entry.Type().IsRegular() {
					continue
				}
			}
			path := filepath.Join(sourceDir, entry.Name())
			regular, err := regularFileWithin(sourceDir, entry.Name())
			if err != nil {
				return err
			}
			if regular {
				if err := copyOne(entry.Name(), path); err != nil {
					return err
				}
			}
		}
	}
	for _, file := range files {
		if err := copyOne(file.name, file.path); err != nil {
			return err
		}
	}

	if err := managed.removeTree(target); err != nil {
		return err
	}
	for name, file := range contents {
		if err := managed.writeSecret(filepath.Join(target, name), file.data); err != nil {
			return err
		}
	}
	if err := rehashDirectory(target); err != nil {
		return fmt.Errorf("rehash CA directory %q: %w", target, err)
	}
	return nil
}

func generatedProviderFileName(variable, path, defaultExtension string) string {
	extension := filepath.Ext(path)
	if extension == "" {
		extension = defaultExtension
	}
	return variable + extension
}

func generatedProviderCAFileName(variable, path string) string {
	extension := strings.ToLower(filepath.Ext(path))
	switch extension {
	case ".pem", ".crt", ".cer", ".crl":
	default:
		extension = ".pem"
	}
	return variable + extension
}

func resolveSecretDirectory(name, current, legacy string) (string, error) {
	currentUsed, err := directoryHasFiles(current)
	if err != nil {
		return "", err
	}
	legacyUsed := false
	if legacy != "" {
		legacyUsed, err = directoryHasFiles(legacy)
		if err != nil {
			return "", err
		}
	}
	if currentUsed && legacyUsed {
		return "", fmt.Errorf("%s: both %q and deprecated location %q contain files", name, current, legacy)
	}
	if currentUsed {
		return current, nil
	}
	if legacyUsed {
		LogWarn("%s uses deprecated location %s. Please migrate to %s", name, legacy, current)
		return legacy, nil
	}
	return current, nil
}

func directoryHasFiles(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read secret directory %q: %w", path, err)
	}
	for _, entry := range entries {
		// Kubernetes projected volumes contain bookkeeping symlinks. Count
		// only entries that resolve to a regular file.
		regular, err := regularFileWithin(path, entry.Name())
		if err != nil {
			return false, err
		}
		if regular {
			return true, nil
		}
	}
	return false, nil
}
