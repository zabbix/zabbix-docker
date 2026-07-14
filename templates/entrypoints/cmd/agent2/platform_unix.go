//go:build !windows

package main

const (
	agent2Binary     = "/usr/sbin/zabbix_agent2"
	nvidiaCommand    = "nvidia-smi"
	pluginExecSuffix = ""
)

func pluginBinDir(homeDir string) string {
	return "/usr/sbin/zabbix-agent2-plugin"
}
