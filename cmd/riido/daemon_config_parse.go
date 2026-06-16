package main

import (
	"net"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

const defaultDevelopmentPprofAddr = "127.0.0.1:6061"

func defaultDaemonID(configuredDaemonID, deviceID string) string {
	if daemonID := strings.TrimSpace(configuredDaemonID); daemonID != "" {
		return daemonID
	}
	if devicePrincipalID := strings.TrimSpace(deviceID); devicePrincipalID != "" {
		return devicePrincipalID
	}
	return "agentd-local"
}

func parseOptionalNonNegativeInt(raw, name string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, daemonWrapf(ErrDaemonConfig, "settings.parse-non-negative-int", err, "%s must be a non-negative integer", name)
	}
	return n, nil
}

func parseOptionalDurationSeconds(raw, name string) (time.Duration, error) {
	n, err := parseOptionalNonNegativeInt(raw, name)
	if err != nil {
		return 0, err
	}
	return time.Duration(n) * time.Second, nil
}

func parseOptionalPositiveDurationSeconds(raw, name string, fallback time.Duration) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, daemonWrapf(ErrDaemonConfig, "settings.parse-positive-int", err, "%s must be a positive integer", name)
	}
	return time.Duration(n) * time.Second, nil
}

func parseDaemonPprofAddr(raw, profile string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if daemonProfileEnablesPprof(profile) {
			return defaultDevelopmentPprofAddr, nil
		}
		return "", nil
	}
	switch strings.ToLower(raw) {
	case "0", "false", "off", "disabled":
		return "", nil
	}
	host, port, err := net.SplitHostPort(raw)
	if err != nil {
		if strings.HasPrefix(raw, ":") {
			host = "127.0.0.1"
			port = strings.TrimPrefix(raw, ":")
		} else {
			return "", daemonWrapf(ErrDaemonConfig, "settings.parse-pprof-addr", err, "%s must be a loopback host:port", envDaemonPprofAddr)
		}
	}
	if strings.TrimSpace(port) == "" {
		return "", daemonErrorf(ErrDaemonConfig, "settings.parse-pprof-addr", "%s requires a port", envDaemonPprofAddr)
	}
	host = strings.TrimSpace(host)
	if host == "" {
		host = "127.0.0.1"
	}
	if !daemonPprofHostIsLoopback(host) {
		return "", daemonErrorf(ErrDaemonConfig, "settings.validate-pprof-addr", "%s must bind to localhost or a loopback address", envDaemonPprofAddr)
	}
	return net.JoinHostPort(host, port), nil
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

func parseRuntimeAgents(raw string) []runtimeactor.AgentStatus {
	parts := strings.Split(raw, ",")
	out := make([]runtimeactor.AgentStatus, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		out = append(out, runtimeactor.AgentStatus{
			AgentID: slugAgentName(name),
			Name:    name,
			State:   "online",
		})
	}
	return out
}

func slugAgentName(name string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
