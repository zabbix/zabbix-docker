package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func prepareHistoryProviderTLSFiles(managed *managedSecrets, env Environment, files []fileSecretSpec) error {
	groups := []struct {
		suffix, location, category, extension string
	}{
		{suffix: "_SSL_CERT", location: "ZBX_SSLCERTLOCATION", category: "ssl/certs", extension: ".pem"},
		{suffix: "_SSL_KEY", location: "ZBX_SSLKEYLOCATION", category: "ssl/keys", extension: ".key"},
	}

	for _, group := range groups {
		var specs []fileSecretSpec
		for _, spec := range files {
			if strings.HasPrefix(spec.inline, "ZBX_HISTORYPROVIDER_") && strings.HasSuffix(spec.inline, group.suffix) && env[spec.file] != "" {
				specs = append(specs, spec)
			}
		}
		if len(specs) == 0 {
			continue
		}
		if err := prepareHistoryProviderFileGroup(managed, env, group.location, group.category, group.extension, specs); err != nil {
			return err
		}
	}
	return nil
}

func prepareHistoryProviderFileGroup(managed *managedSecrets, env Environment, location, category, extension string,
	specs []fileSecretSpec) error {
	type content struct {
		data   []byte
		source string
	}

	sourceDir := env[location]
	targetDir := filepath.Join(internalSecretsRoot, filepath.FromSlash(category))
	contents := make(map[string]content)

	entries, err := os.ReadDir(sourceDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read TLS source directory %q: %w", sourceDir, err)
	}
	for _, entry := range entries {
		if filepath.Clean(sourceDir) == filepath.Clean(targetDir) && strings.HasPrefix(entry.Name(), "ZBX_HISTORYPROVIDER_") {
			continue
		}
		regular, err := regularFileWithin(sourceDir, entry.Name())
		if err != nil {
			return err
		}
		if !regular {
			continue
		}
		path := filepath.Join(sourceDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read TLS file %q: %w", path, err)
		}
		contents[entry.Name()] = content{data: data, source: path}
	}

	outputNames := make(map[string]string, len(specs))
	for _, spec := range specs {
		path := env[spec.file]
		name := generatedProviderFileName(spec.file, path, extension)
		if previous, exists := contents[name]; exists {
			return fmt.Errorf("TLS file name %q is supplied by both %q and %q", name, previous.source, path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("%s: read secret file %q: %w", spec.file, path, err)
		}
		contents[name] = content{data: data, source: path}
		outputNames[spec.file] = name
	}

	if err := managed.removeTree(targetDir); err != nil {
		return err
	}
	for name, file := range contents {
		if err := managed.writeSecret(filepath.Join(targetDir, name), file.data); err != nil {
			return err
		}
	}

	for _, file := range vaultTLSFileDefinitions {
		if file.location != location || env[file.file] == "" {
			continue
		}
		relative, err := filepath.Rel(sourceDir, env[file.file])
		if err == nil && filepath.IsLocal(relative) {
			env[file.file] = filepath.Join(targetDir, relative)
		}
	}
	for file, name := range outputNames {
		env[file] = name
	}
	env[location] = targetDir
	return nil
}
