package bootstrap

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Hostname returns the container hostname. With fqdn it additionally tries
// to resolve the fully qualified name via DNS, falling back to the short
// name when resolution fails.
func Hostname(fqdn bool) (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("determine hostname: %w", err)
	}

	if !fqdn || strings.Contains(hostname, ".") {
		return hostname, nil
	}

	if canonical, err := net.LookupCNAME(hostname); err == nil {
		canonical = strings.TrimSuffix(canonical, ".")
		if strings.Contains(canonical, ".") {
			return canonical, nil
		}
	}

	addresses, _ := net.LookupHost(hostname)
	for _, address := range addresses {
		names, err := net.LookupAddr(address)
		if err == nil && len(names) > 0 {
			name := strings.TrimSuffix(names[0], ".")
			if strings.Contains(name, ".") {
				return name, nil
			}
		}
	}

	return hostname, nil
}

// PrepareCertDir replaces the target directory contents with the CA
// certificates from source and creates OpenSSL hash links in target.
func PrepareCertDir(sourceDir, targetDir string) error {
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return fmt.Errorf("read certificate directory %q: %w", targetDir, err)
	}

	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(targetDir, entry.Name())); err != nil {
			return fmt.Errorf("clear certificate directory %q: %w", targetDir, err)
		}
	}

	if err := os.CopyFS(targetDir, os.DirFS(sourceDir)); err != nil {
		return fmt.Errorf("copy certificates from %q to %q: %w", sourceDir, targetDir, err)
	}

	command := exec.Command("openssl", "rehash", targetDir)
	command.Stdout = io.Discard
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		LogWarn("openssl rehash failed for '%s': %v", targetDir, err)
	}

	return nil
}
