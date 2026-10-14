package main

import (
	"strings"

	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

type daemonIdentityEnv struct {
	deviceName string
	profile    string
	owner      string
}

func loadDaemonIdentityEnv(getenv func(string) string, hostname func() (string, error)) daemonIdentityEnv {
	return daemonIdentityEnv{
		deviceName: defaultDaemonDeviceName(getenv, hostname),
		profile:    textutil.Default(getenv(envDaemonProfile), "local"),
		owner:      defaultRuntimeOwner(getenv),
	}
}

func defaultDaemonDeviceName(getenv func(string) string, hostname func() (string, error)) string {
	deviceName := strings.TrimSpace(getenv(envDeviceName))
	if deviceName == "" {
		if host, err := hostname(); err == nil {
			deviceName = strings.TrimSpace(host)
		}
	}
	if deviceName == "" {
		return "localhost"
	}
	return deviceName
}

func defaultRuntimeOwner(getenv func(string) string) string {
	owner := strings.TrimSpace(getenv(envRuntimeOwner))
	if owner == "" {
		owner = strings.TrimSpace(getenv("USER"))
	}
	if owner == "" {
		return "local"
	}
	return owner
}
