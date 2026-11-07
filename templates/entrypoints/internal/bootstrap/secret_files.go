package bootstrap

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func cleanupStaleProviderSecrets(managed *managedSecrets, component SecretComponent, command SecretCommand) error {
	families := make([]string, 0, 2)
	switch command {
	case SecretsInitHistory:
		families = append(families, "history")
	case SecretsInitTelemetry:
		families = append(families, "telemetry")
	case SecretsService:
		if isServer(component) || isWeb(component) {
			families = append(families, "history", "telemetry")
		} else if isProxy(component) {
			families = append(families, "telemetry")
		}
	}
	if len(families) == 0 {
		return nil
	}
	root, err := managed.openExisting()
	if err != nil {
		return err
	}
	if root == nil {
		return nil
	}
	providers, err := root.Open("providers")
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open managed provider directory: %w", err)
	}
	entries, err := providers.ReadDir(-1)
	providers.Close()
	if err != nil {
		return fmt.Errorf("read managed provider directory: %w", err)
	}
	for _, family := range families {
		for _, provider := range entries {
			if !provider.IsDir() || !strings.HasPrefix(provider.Name(), family+"-") {
				continue
			}
			directory := filepath.Join("providers", provider.Name())
			index := strings.TrimPrefix(provider.Name(), family+"-")
			prefix := "ZBX_" + strings.ToUpper(family) + "PROVIDER_" + index
			for _, secret := range providerSecretDefinitions {
				if secret.Kind != ProviderFileSecret {
					continue
				}
				if err := managed.removeSecretRelative(filepath.Join(directory, prefix+secret.Suffix+"FILE")); err != nil {
					return err
				}
			}
			caDir := filepath.Join(directory, "ca")
			ca, readErr := root.Open(caDir)
			if errors.Is(readErr, os.ErrNotExist) {
				continue
			}
			if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
				return fmt.Errorf("read stale provider CA directory %q: %w", caDir, readErr)
			}
			caEntries, readErr := ca.ReadDir(-1)
			ca.Close()
			if readErr != nil {
				return fmt.Errorf("read stale provider CA directory %q: %w", caDir, readErr)
			}
			for _, entry := range caEntries {
				if err := root.RemoveAll(filepath.Join(caDir, entry.Name())); err != nil {
					return fmt.Errorf("remove stale provider CA entry %q: %w", entry.Name(), err)
				}
			}
		}
	}
	return nil
}

func normalizeSecretAlias(env Environment, alias secretAlias) error {
	legacy := env[alias.legacy]
	canonical := env[alias.canonical]
	if legacy == "" {
		delete(env, alias.legacy)
		return nil
	}
	if canonical != "" {
		if legacy != canonical {
			return fmt.Errorf("%s and %s specify different values; remove deprecated variable %s",
				alias.legacy, alias.canonical, alias.legacy)
		}
		LogWarn("%s is deprecated and ignored because %s is set to the same value", alias.legacy, alias.canonical)
		delete(env, alias.legacy)
		return nil
	}
	env[alias.canonical] = legacy
	LogWarn("%s is deprecated; use %s", alias.legacy, alias.canonical)
	delete(env, alias.legacy)
	return nil
}

func normalizeEmptySecretSources(env Environment, files []fileSecretSpec, values []valueSecretSpec) {
	removeEmpty := func(name string) {
		if value, exists := env[name]; exists && value == "" {
			delete(env, name)
		}
	}
	for _, spec := range files {
		removeEmpty(spec.inline)
		removeEmpty(spec.file)
	}
	for _, spec := range values {
		removeEmpty(spec.name)
		removeEmpty(spec.name + "_FILE")
	}
}

func resolveFileSecret(managed *managedSecrets, env Environment, spec fileSecretSpec) error {
	inline := env[spec.inline]
	path := env[spec.file]
	managedPath := filepath.Join(internalSecretsRoot, filepath.FromSlash(spec.category), spec.file)
	delete(env, spec.inline)
	if !spec.active {
		if err := managed.removeSecret(managedPath); err != nil {
			return err
		}
		delete(env, spec.file)
		return nil
	}

	if inline == "" && path == "" {
		if err := managed.removeSecret(managedPath); err != nil {
			return err
		}
		delete(env, spec.file)
		return nil
	}
	if inline != "" && path != "" {
		return fmt.Errorf("both variables %s and %s are set (but are exclusive)", spec.inline, spec.file)
	}

	if inline != "" {
		if err := managed.writeSecret(managedPath, []byte(inline)); err != nil {
			return fmt.Errorf("write %s: %w", spec.file, err)
		}
		env[spec.file] = managedPath
		return nil
	}

	resolved, err := locateSecret(spec.file, path, spec.category, spec.standard, spec.legacyDir)
	if err != nil {
		return err
	}
	env[spec.file] = resolved
	return nil
}

func resolveValueSecret(env Environment, spec valueSecretSpec) error {
	fileName := spec.name + "_FILE"
	value := env[spec.name]
	path := env[fileName]
	if !spec.active {
		delete(env, spec.name)
		delete(env, fileName)
		return nil
	}
	if value != "" && path != "" {
		return fmt.Errorf("both variables %s and %s are set (but are exclusive)", spec.name, fileName)
	}
	if path == "" {
		delete(env, fileName)
		return nil
	}

	resolved, err := locateSecret(fileName, path, spec.category, true, "")
	if err != nil {
		return err
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		return fmt.Errorf("%s: read secret file %q: %w", fileName, resolved, err)
	}
	env[spec.name] = strings.TrimRight(string(data), "\r\n")
	delete(env, fileName)
	return nil
}

func locateSecret(variable, path, category string, standard bool, legacyDir string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	if !filepath.IsLocal(path) {
		return "", fmt.Errorf("%s: relative path %q escapes its secret directory", variable, path)
	}

	newBase := filepath.Join(externalSecretsRoot, filepath.FromSlash(category))
	newPath := filepath.Join(newBase, path)
	if standard {
		standardPath := filepath.Join(dockerSecretsRoot, path)
		// /run/secrets is the standard Docker secrets contract and wins when
		// both locations contain the same relative name.
		standardExists, err := regularFileWithin(dockerSecretsRoot, path)
		if err != nil {
			return "", fmt.Errorf("%s: %w", variable, err)
		}
		newExists, err := regularFileWithin(newBase, path)
		if err != nil {
			return "", fmt.Errorf("%s: %w", variable, err)
		}
		if standardExists {
			if newExists {
				LogWarn("%s exists in both %s and %s; using standard Docker secret %s", variable, standardPath, newPath, standardPath)
			}
			return standardPath, nil
		}
		if newExists {
			return newPath, nil
		}
		return "", fmt.Errorf("%s: secret file %q was not found (checked %q and %q)", variable, path, standardPath, newPath)
	}

	legacyPath := ""
	if legacyDir != "" {
		legacyPath = filepath.Join(legacyDir, path)
	}
	newExists, err := regularFileWithin(newBase, path)
	if err != nil {
		return "", fmt.Errorf("%s: %w", variable, err)
	}
	legacyExists := false
	if legacyPath != "" {
		legacyExists, err = regularFileWithin(legacyDir, path)
		if err != nil {
			return "", fmt.Errorf("%s: %w", variable, err)
		}
	}
	if newExists && legacyExists {
		return "", fmt.Errorf("%s: secret file exists in both %q and deprecated location %q", variable, newPath, legacyPath)
	}
	if newExists {
		return newPath, nil
	}
	if legacyExists {
		LogWarn("%s resolved from deprecated location %s. Please migrate to %s", variable, legacyPath, newPath)
		return legacyPath, nil
	}
	if legacyPath == "" {
		return "", fmt.Errorf("%s: secret file %q was not found (checked %q)", variable, path, newPath)
	}
	return "", fmt.Errorf("%s: secret file %q was not found (checked %q and %q)", variable, path, newPath, legacyPath)
}

// ResolveExternalSecretFile resolves a file secret from the current external
// secrets hierarchy. Relative paths must name an existing regular file within
// the requested category; absolute paths are preserved as user-managed paths.
func ResolveExternalSecretFile(variable, path, category string) (string, error) {
	return locateSecret(variable, path, category, false, "")
}

func (managed *managedSecrets) writeSecret(path string, data []byte) error {
	relative, err := managedRelativePath(path)
	if err != nil {
		return err
	}
	root, err := managed.openWritable()
	if err != nil {
		return err
	}
	dir := filepath.Dir(relative)
	if err := root.MkdirAll(dir, 0o700); err != nil {
		return writableSecretDirError(filepath.Join(internalSecretsRoot, dir), err)
	}
	if err := root.Chmod(dir, 0o700); err != nil {
		return writableSecretDirError(filepath.Join(internalSecretsRoot, dir), err)
	}
	temporaryName, temporary, err := createManagedTemp(root, dir)
	if err != nil {
		return writableSecretDirError(filepath.Join(internalSecretsRoot, dir), err)
	}
	defer root.Remove(temporaryName)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return root.Rename(temporaryName, relative)
}

func createManagedTemp(root *os.Root, directory string) (string, *os.File, error) {
	for range 10 {
		var token [8]byte
		if _, err := rand.Read(token[:]); err != nil {
			return "", nil, err
		}
		name := filepath.Join(directory, fmt.Sprintf(".secret-%x", token))
		file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		return name, file, err
	}
	return "", nil, errors.New("cannot allocate a unique temporary secret file")
}

type managedSecrets struct {
	root *os.Root
}

func (managed *managedSecrets) open(create bool) (*os.Root, error) {
	if managed.root != nil {
		return managed.root, nil
	}
	if create {
		if err := os.MkdirAll(internalSecretsRoot, 0o700); err != nil {
			return nil, writableSecretDirError(internalSecretsRoot, err)
		}
	}
	info, err := os.Lstat(internalSecretsRoot)
	if errors.Is(err, os.ErrNotExist) && !create {
		return nil, nil
	}
	if err != nil {
		return nil, writableSecretDirError(internalSecretsRoot, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return nil, fmt.Errorf("generated secret root %q must be a real directory", internalSecretsRoot)
	}
	root, err := os.OpenRoot(internalSecretsRoot)
	if err != nil {
		return nil, writableSecretDirError(internalSecretsRoot, err)
	}
	managed.root = root
	return root, nil
}

func (managed *managedSecrets) openExisting() (*os.Root, error) {
	return managed.open(false)
}

func (managed *managedSecrets) openWritable() (*os.Root, error) {
	return managed.open(true)
}

func (managed *managedSecrets) close() {
	if managed.root != nil {
		managed.root.Close()
		managed.root = nil
	}
}

func managedRelativePath(path string) (string, error) {
	absoluteRoot, err := filepath.Abs(internalSecretsRoot)
	if err != nil {
		return "", err
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(absoluteRoot, absolutePath)
	if err != nil || !filepath.IsLocal(relative) {
		return "", fmt.Errorf("generated secret path %q escapes %q", path, internalSecretsRoot)
	}
	return relative, nil
}

func writableSecretDirError(dir string, err error) error {
	return fmt.Errorf("prepare generated secret directory %q: %w; mount a writable tmpfs at %s", dir, err, internalSecretsRoot)
}

func (managed *managedSecrets) removeSecret(path string) error {
	relative, err := managedRelativePath(path)
	if err != nil {
		return err
	}
	return managed.removeSecretRelative(relative)
}

func (managed *managedSecrets) removeTree(path string) error {
	relative, err := managedRelativePath(path)
	if err != nil {
		return err
	}
	root, err := managed.openExisting()
	if err != nil {
		return err
	}
	if root == nil {
		return nil
	}
	if err := root.RemoveAll(relative); err != nil {
		return fmt.Errorf("remove stale generated path %q: %w", relative, err)
	}
	return nil
}

func (managed *managedSecrets) removeSecretRelative(relative string) error {
	root, err := managed.openExisting()
	if err != nil {
		return err
	}
	if root == nil {
		return nil
	}
	if err := root.Remove(relative); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("remove stale generated secret %q: %w", relative, err)
	}
	return nil
}

func regularFileWithin(base, relative string) (bool, error) {
	if !filepath.IsLocal(relative) {
		return false, fmt.Errorf("relative path %q escapes %q", relative, base)
	}
	candidate := filepath.Join(base, relative)
	info, err := os.Stat(candidate)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("access secret file %q: %w", candidate, err)
	}
	if !info.Mode().IsRegular() {
		return false, nil
	}

	resolvedBase, err := filepath.EvalSymlinks(base)
	if errors.Is(err, os.ErrNotExist) {
		resolvedBase, err = filepath.Abs(base)
	}
	if err != nil {
		return false, fmt.Errorf("resolve secret directory %q: %w", base, err)
	}
	resolvedCandidate, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return false, fmt.Errorf("resolve secret file %q: %w", candidate, err)
	}
	relativeResolved, err := filepath.Rel(resolvedBase, resolvedCandidate)
	if err != nil || !filepath.IsLocal(relativeResolved) {
		return false, fmt.Errorf("secret file %q resolves outside %q", candidate, base)
	}
	return true, nil
}

// IsGeneratedSecretPath reports whether path belongs to entrypoint-managed
// secrets_internal output.
func IsGeneratedSecretPath(path string) bool {
	clean := filepath.Clean(path)
	root := filepath.Clean(internalSecretsRoot)
	return clean == root || strings.HasPrefix(clean, root+string(filepath.Separator))
}
