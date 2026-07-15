package main

import (
	"path/filepath"
)

const (
	agent2Binary     = `C:\zabbix\sbin\zabbix_agent2.exe`
	nvidiaCommand    = "nvidia-smi.exe"
	pluginExecSuffix = ".exe"
)

func pluginBinDir(homeDir string) string {
	return filepath.Join(homeDir, "zabbix-agent2-plugin")
}
