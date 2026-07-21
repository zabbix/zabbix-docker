package bootstrap

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
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

// RehashCertificateDirectory creates OpenSSL hash links for the CA
// certificates in directory. It is best effort: failures are logged and do
// not stop container startup. A missing directory is silently ignored.
func RehashCertificateDirectory(directory string) {
	if directory == "" {
		return
	}

	info, err := os.Stat(directory)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		LogWarn("cannot inspect certificate directory '%s': %v", directory, err)
		return
	}
	if !info.IsDir() {
		LogWarn("cannot rehash certificate directory '%s': not a directory", directory)
		return
	}

	openssl, err := exec.LookPath("openssl")
	if err != nil {
		LogWarn("cannot rehash certificate directory '%s': %v", directory, err)
		return
	}

	command := exec.Command(openssl, "rehash", directory)
	command.Stdout = io.Discard
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		LogWarn("openssl rehash failed for '%s': %v", directory, err)
	}
}
