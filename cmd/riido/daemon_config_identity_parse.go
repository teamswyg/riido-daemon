package main

import "strings"

func defaultDaemonID(configuredDaemonID, deviceID string) string {
	if daemonID := strings.TrimSpace(configuredDaemonID); daemonID != "" {
		return daemonID
	}
	if devicePrincipalID := strings.TrimSpace(deviceID); devicePrincipalID != "" {
		return devicePrincipalID
	}
	return "agentd-local"
}
