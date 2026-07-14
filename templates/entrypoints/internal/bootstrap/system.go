package bootstrap

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
)

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

func RehashCertificateDirectory(directory string) error {
	if directory == "" {
		return nil
	}

	info, err := os.Stat(directory)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil || !info.IsDir() {
		return nil
	}

	openssl, err := exec.LookPath("openssl")
	if err != nil {
		return nil
	}

	command := exec.Command(openssl, "rehash", directory)
	command.Stdout = io.Discard
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		LogWarn("openssl rehash failed for '%s'", directory)
	}

	return nil
}
