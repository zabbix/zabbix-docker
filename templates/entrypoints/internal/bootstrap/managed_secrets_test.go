package bootstrap

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestManagedSecretsOpensRootLazilyOnce(t *testing.T) {
	previous := internalSecretsRoot
	internalSecretsRoot = filepath.Join(t.TempDir(), "internal")
	t.Cleanup(func() { internalSecretsRoot = previous })

	managed := &managedSecrets{}
	defer managed.close()

	root, err := managed.openExisting()
	if err != nil {
		t.Fatal(err)
	}
	if root != nil {
		t.Fatal("openExisting() created a missing managed root")
	}
	if _, err := os.Stat(internalSecretsRoot); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("managed root exists before first write: %v", err)
	}

	writable, err := managed.openWritable()
	if err != nil {
		t.Fatal(err)
	}
	reused, err := managed.openExisting()
	if err != nil {
		t.Fatal(err)
	}
	if reused != writable {
		t.Fatal("managed root was reopened instead of reused")
	}
}

func TestResolveSecretsWithExternalValueDoesNotCreateManagedRoot(t *testing.T) {
	root := t.TempDir()
	previousExternal, previousInternal, previousDocker := externalSecretsRoot, internalSecretsRoot, dockerSecretsRoot
	externalSecretsRoot = filepath.Join(root, "external")
	internalSecretsRoot = filepath.Join(root, "internal")
	dockerSecretsRoot = filepath.Join(root, "docker")
	t.Cleanup(func() {
		externalSecretsRoot, internalSecretsRoot, dockerSecretsRoot = previousExternal, previousInternal, previousDocker
	})
	if err := os.MkdirAll(dockerSecretsRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	writeTestSecret(t, filepath.Join(dockerSecretsRoot, "password"), "secret")

	env := Environment{
		"ZABBIX_USER_HOME_DIR": t.TempDir(),
		"MYSQL_PASSWORD_FILE":  "password",
	}
	if err := ResolveSecrets(env, SecretsServerMySQL, SecretsInitDB); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(internalSecretsRoot); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("external-only resolution created managed root: %v", err)
	}
}
