//go:build !windows

package bootstrap

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
)

func ConfigureIdentity(env Environment) error {
	uid := os.Getuid()
	if uid == 0 {
		env["ZBX_ALLOWROOT"] = "1"
		return nil
	}

	account, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		return fmt.Errorf("lookup user for uid %d: %w", uid, err)
	}

	env["ZBX_USER"] = account.Username
	return nil
}
