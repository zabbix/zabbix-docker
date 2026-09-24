//go:build !windows

package bootstrap

import (
	"os"
	"os/user"
	"strconv"
)

// ConfigureRunUser tells Zabbix how to run under the current UID: root
// gets ZBX_ALLOWROOT, any other user is looked up and exported as ZBX_USER.
func ConfigureRunUser(env Environment) error {
	uid := os.Getuid()
	if uid == 0 {
		env["ZBX_ALLOWROOT"] = "1"
		return nil
	}

	account, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		// Containers may run with an arbitrary UID that has no passwd entry
		// (for example, the random UID assigned by OpenShift). Zabbix ignores
		// User when it is already running as a non-root user, so keep the UID
		// instead of failing startup.
		env["ZBX_USER"] = strconv.Itoa(uid)
		return nil
	}

	env["ZBX_USER"] = account.Username
	return nil
}
