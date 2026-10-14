//go:build !windows

package main

import (
	"slices"
	"strings"
)

func daemonCommandLineLooksLikeForegroundStart(fields []string) bool {
	if len(fields) == 0 {
		return false
	}
	for i, field := range fields {
		if field != string(mainCommandDaemon) {
			continue
		}
		if daemonCommandFieldsAfterDaemonLookLikeForegroundStart(fields[i+1:]) {
			return true
		}
	}
	return false
}

func daemonCommandFieldsAfterDaemonLookLikeForegroundStart(fields []string) bool {
	return len(fields) > 0 &&
		fields[0] == string(daemonCommandStart) &&
		fieldsAfterContain(fields[1:], "--foreground")
}

func fieldsAfterContain(fields []string, want string) bool {
	return slices.Contains(fields, want)
}

func daemonCommandFieldsContainSocket(fields []string, socket string) bool {
	for i, field := range fields {
		if field == "--socket" {
			return i+1 < len(fields) && fields[i+1] == socket
		}
		if value, ok := strings.CutPrefix(field, "--socket="); ok {
			return value == socket
		}
	}
	return false
}
