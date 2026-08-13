package postgresql

import (
	"fmt"
	"strconv"
	"strings"
)

const socketDir = "/var/run/postgresql"

type endpoint struct {
	host string
	port string
}

func parseEndpoints(hosts, defaultPort string) ([]endpoint, error) {
	if hosts == "" {
		return []endpoint{{host: socketDir, port: defaultPort}}, nil
	}

	parts := strings.Split(hosts, ",")
	endpoints := make([]endpoint, 0, len(parts))
	for i, raw := range parts {
		value := strings.TrimSpace(raw)
		if value == "" {
			return nil, fmt.Errorf("endpoint %d is empty", i+1)
		}

		host, port, err := parseEndpoint(value, defaultPort)
		if err != nil {
			return nil, fmt.Errorf("endpoint %d %q: %w", i+1, value, err)
		}
		endpoints = append(endpoints, endpoint{host: host, port: port})
	}

	return endpoints, nil
}

func parseEndpoint(value, defaultPort string) (string, string, error) {
	if strings.HasPrefix(value, "[") {
		bracketEnd := strings.IndexByte(value, ']')
		if bracketEnd == -1 {
			return "", "", fmt.Errorf("missing closing bracket in IPv6 address")
		}

		host := strings.TrimSpace(value[1:bracketEnd])
		if host == "" {
			return "", "", fmt.Errorf("host is empty")
		}

		suffix := strings.TrimSpace(value[bracketEnd+1:])
		if suffix == "" {
			return host, defaultPort, nil
		}
		if !strings.HasPrefix(suffix, ":") {
			return "", "", fmt.Errorf("unexpected data after IPv6 address")
		}

		port := strings.TrimSpace(suffix[1:])
		if err := validatePort(port); err != nil {
			return "", "", err
		}

		return host, port, nil
	}

	colons := strings.Count(value, ":")
	if colons == 0 || colons > 1 {
		return value, defaultPort, nil
	}

	host, port, _ := strings.Cut(value, ":")
	host = strings.TrimSpace(host)
	port = strings.TrimSpace(port)
	if host == "" {
		return "", "", fmt.Errorf("host is empty")
	}
	if err := validatePort(port); err != nil {
		return "", "", err
	}

	return host, port, nil
}

func validatePort(port string) error {
	if port == "" {
		return fmt.Errorf("port is empty")
	}
	if _, err := strconv.ParseUint(port, 10, 16); err != nil {
		return fmt.Errorf("invalid port %q: %w", port, err)
	}

	return nil
}
