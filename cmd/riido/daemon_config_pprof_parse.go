package main

import (
	"net"
	"strings"
)

const defaultDevelopmentPprofAddr = "127.0.0.1:6061"

func parseDaemonPprofAddr(raw, profile string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if daemonProfileEnablesPprof(profile) {
			return defaultDevelopmentPprofAddr, nil
		}
		return "", nil
	}
	if daemonPprofAddrDisabled(raw) {
		return "", nil
	}
	host, port, err := splitDaemonPprofHostPort(raw)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(port) == "" {
		return "", daemonErrorf(ErrDaemonConfig, "settings.parse-pprof-addr", "%s requires a port", envDaemonPprofAddr)
	}
	if !daemonPprofHostIsLoopback(host) {
		return "", daemonErrorf(ErrDaemonConfig, "settings.validate-pprof-addr", "%s must bind to localhost or a loopback address", envDaemonPprofAddr)
	}
	return net.JoinHostPort(host, port), nil
}

func daemonPprofAddrDisabled(raw string) bool {
	switch strings.ToLower(raw) {
	case "0", "false", "off", "disabled":
		return true
	default:
		return false
	}
}
