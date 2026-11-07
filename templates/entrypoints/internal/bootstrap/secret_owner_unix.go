//go:build !windows

package bootstrap

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"strconv"
)

// ChownGeneratedSecrets transfers the generated secret tree to a service
// account when the entrypoint runs as root.
func ChownGeneratedSecrets(username string) error {
	if os.Getuid() != 0 || username == "" {
		return nil
	}
	account, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("look up generated-secret owner %q: %w", username, err)
	}
	uid, err := strconv.Atoi(account.Uid)
	if err != nil {
		return fmt.Errorf("parse uid for %q: %w", username, err)
	}
	gid, err := strconv.Atoi(account.Gid)
	if err != nil {
		return fmt.Errorf("parse gid for %q: %w", username, err)
	}
	managed := &managedSecrets{}
	defer managed.close()
	root, err := managed.openExisting()
	if err != nil {
		return err
	}
	if root == nil {
		return nil
	}
	if err := chownGeneratedSecretTree(root, uid, gid); err != nil {
		return fmt.Errorf("change generated secret tree owner: %w", err)
	}
	return nil
}

func chownGeneratedSecretTree(root *os.Root, uid, gid int) error {
	if err := fs.WalkDir(root.FS(), ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		return root.Chown(path, uid, gid)
	}); err != nil {
		return err
	}
	return root.Chown(".", uid, gid)
}
