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

func resolveGlobalCA(managed *managedSecrets, env Environment) error {
	if env["ZBX_SSLCALOCATION"] != "" {
		return nil
	}
	home := env["ZABBIX_USER_HOME_DIR"]
	source, err := resolveSecretDirectory(
		"ZBX_SSLCALOCATION",
		filepath.Join(externalSecretsRoot, "ssl", "ca"),
		filepath.Join(home, "ssl", "ssl_ca"),
	)
	if err != nil {
		return err
	}
	target := filepath.Join(internalSecretsRoot, "ssl", "ca")
	if err := prepareManagedCADir(managed, target, []string{source}, nil); err != nil {
		return err
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
	if !currentUsed && env["ZBX_APM_CAFILE"] == "" {
		delete(env, "ZBX_APM_CA_LOCATION")
		return managed.removeTree(filepath.Join(internalSecretsRoot, "apm", "ca"))
	}
	source := current
	var extra []string
	if path := env["ZBX_APM_CAFILE"]; path != "" {
		extra = append(extra, path)
	}
	target := filepath.Join(internalSecretsRoot, "apm", "ca")
	if err := prepareManagedCADir(managed, target, []string{source}, extra); err != nil {
		return err
	}
	env["ZBX_APM_CA_LOCATION"] = target
	return nil
}

func resolveProviderCAs(managed *managedSecrets, env Environment, files []fileSecretSpec) error {
	for _, spec := range files {
		if !strings.HasSuffix(spec.inline, "_SSL_CA") {
			continue
		}
		path := env[spec.file]
		if path == "" {
			continue
		}
		target := filepath.Join(internalSecretsRoot, filepath.FromSlash(spec.category), "ca")
		if err := prepareManagedCADir(managed, target, nil, []string{path}); err != nil {
			return fmt.Errorf("%s: %w", spec.file, err)
		}
		env[strings.TrimSuffix(spec.file, "FILE")+"LOCATION"] = target
	}
	return nil
}

func prepareManagedCADir(managed *managedSecrets, target string, sources, files []string) error {
	relativeTarget, err := managedRelativePath(target)
	if err != nil {
		return err
	}
	root, err := managed.openWritable()
	if err != nil {
		return err
	}
	if err := root.MkdirAll(relativeTarget, 0o700); err != nil {
		return writableSecretDirError(target, err)
	}
	if err := root.Chmod(relativeTarget, 0o700); err != nil {
		return writableSecretDirError(target, err)
	}
	directory, err := root.Open(relativeTarget)
	if err != nil {
		return fmt.Errorf("read managed CA directory %q: %w", target, err)
	}
	entries, err := directory.ReadDir(-1)
	directory.Close()
	if err != nil {
		return fmt.Errorf("read managed CA directory %q: %w", target, err)
	}
	for _, entry := range entries {
		if err := root.RemoveAll(filepath.Join(relativeTarget, entry.Name())); err != nil {
			return fmt.Errorf("clear managed CA entry %q: %w", filepath.Join(target, entry.Name()), err)
		}
	}

	names := make(map[string]string)
	copyOne := func(source string) error {
		if source == "" {
			return nil
		}
		name := filepath.Base(source)
		if previous, exists := names[name]; exists {
			return fmt.Errorf("CA file name %q is supplied by both %q and %q", name, previous, source)
		}
		data, err := os.ReadFile(source)
		if err != nil {
			return fmt.Errorf("read CA file %q: %w", source, err)
		}
		if err := managed.writeSecret(filepath.Join(target, name), data); err != nil {
			return err
		}
		names[name] = source
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
			path := filepath.Join(sourceDir, entry.Name())
			regular, err := regularFileWithin(sourceDir, entry.Name())
			if err != nil {
				return err
			}
			if regular {
				if err := copyOne(path); err != nil {
					return err
				}
			}
		}
	}
	for _, path := range files {
		if err := copyOne(path); err != nil {
			return err
		}
	}
	if err := rehashDirectory(target); err != nil {
		return fmt.Errorf("rehash CA directory %q: %w", target, err)
	}
	return nil
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
