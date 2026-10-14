//go:build !windows

package main

import (
	"path/filepath"
	"strings"
)

func daemonCommandLineMatchesPIDIdentity(command string, identity daemonPIDIdentity) bool {
	fields := strings.Fields(command)
	if len(fields) == 0 || !daemonCommandBinaryMatchesPIDIdentity(fields[0], identity) {
		return false
	}
	if !daemonCommandLineLooksLikeForegroundStart(fields) {
		return false
	}
	socket := strings.TrimSpace(identity.Socket)
	if socket == "" {
		return false
	}
	return daemonCommandFieldsContainSocket(fields, socket)
}

func daemonCommandBinaryMatchesPIDIdentity(argv0 string, identity daemonPIDIdentity) bool {
	expected := strings.TrimSpace(identity.Executable)
	if expected == "" {
		return daemonCommandBinaryLooksLikeSelf(argv0)
	}
	return filepath.Clean(argv0) == filepath.Clean(expected)
}
