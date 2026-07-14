package main

import (
	"path/filepath"
)

const (
	agent2Binary           = `C:\zabbix\sbin\zabbix_agent2.exe`
	nvidiaCommand          = "nvidia-smi.exe"
	pluginExecutableSuffix = ".exe"
)

func pluginBinaryDirectory(homeDirectory string) string {
	return filepath.Join(homeDirectory, "zabbix-agent2-plugin")
}
