package main

import (
	"net"
	"strings"
)

func splitDaemonPprofHostPort(raw string) (string, string, error) {
	host, port, err := net.SplitHostPort(raw)
	if err != nil {
		if !strings.HasPrefix(raw, ":") {
			return "", "", daemonWrapf(ErrDaemonConfig, "settings.parse-pprof-addr", err, "%s must be a loopback host:port", envDaemonPprofAddr)
		}
		host = "127.0.0.1"
		port = strings.TrimPrefix(raw, ":")
	}
	host = strings.TrimSpace(host)
	if host == "" {
		host = "127.0.0.1"
	}
	return host, port, nil
}

func daemonProfileEnablesPprof(profile string) bool {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case "dev", "development":
		return true
	default:
		return false
	}
}

func daemonPprofHostIsLoopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}
