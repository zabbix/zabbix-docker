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

func rehashDirectory(targetDir string) error {
	command := exec.Command("openssl", "rehash", targetDir)
	command.Stdout = io.Discard
	command.Stderr = os.Stderr
	return command.Run()
}
