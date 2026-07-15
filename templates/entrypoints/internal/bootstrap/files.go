package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrSymlinkSourceNotExist is reported by ReplaceSymlink when the source
// file is missing.
var ErrSymlinkSourceNotExist = errors.New("symlink source does not exist")

// ReplaceSymlink links target to source, replacing an existing file or
// link. The source must be an existing regular file.
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

// RegularFile reports whether path exists and is a regular file.
func RegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

// RemoveLinesContaining deletes every line of the file that contains value.
// A missing file is not an error.
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

// ReplaceInFile applies plain string replacements to the file contents.
// A missing file is not an error.
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

// WriteFilePreservingMode atomically replaces a file while preserving its
// permission bits.
func WriteFilePreservingMode(path string, data []byte) error {
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		return err
	}

	temporary, err := os.CreateTemp(filepath.Dir(resolvedPath), "."+filepath.Base(resolvedPath)+".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	defer temporary.Close()

	if err := temporary.Chmod(info.Mode().Perm()); err != nil {
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}

	return os.Rename(temporary.Name(), resolvedPath)
}
