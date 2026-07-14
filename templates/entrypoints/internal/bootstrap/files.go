package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var ErrSymlinkSourceNotExist = errors.New("symlink source does not exist")

func ReplaceSymlink(source, target string) error {
	info, err := os.Stat(source)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: %s", ErrSymlinkSourceNotExist, source)
		}
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("symlink source %s is not a regular file", source)
	}

	if info, err := os.Lstat(target); err == nil {
		if info.IsDir() {
			return fmt.Errorf("symlink target %s is a directory", target)
		}
		if err := os.Remove(target); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	return os.Symlink(source, target)
}

func RegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func RemoveLinesContaining(path, value string) error {
	if !RegularFile(path) {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	result := lines[:0]
	for _, line := range lines {
		if !strings.Contains(line, value) {
			result = append(result, line)
		}
	}

	return WriteFilePreservingMode(path, []byte(strings.Join(result, "\n")))
}

func ReplaceInFile(path string, replacements map[string]string) error {
	if !RegularFile(path) {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(data)
	for old, replacement := range replacements {
		content = strings.ReplaceAll(content, old, replacement)
	}

	return WriteFilePreservingMode(path, []byte(content))
}

func WriteFilePreservingMode(path string, data []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, info.Mode().Perm())
}
