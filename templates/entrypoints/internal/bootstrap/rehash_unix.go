//go:build !windows

package bootstrap

import (
	"io"
	"os"
	"os/exec"
)

func rehashDirectory(targetDir string) error {
	command := exec.Command("openssl", "rehash", targetDir)
	command.Stdout = io.Discard
	command.Stderr = os.Stderr
	return command.Run()
}
