package main

import (
	"os"
	"sync"

	"github.com/teamswyg/riido-daemon/internal/cliinstallation"
)

// currentInstallChannel is the install channel of the running binary:
// cliinstallation.ChannelStaging for a `riido-staging` install, "" for
// production (RIID-6036). It is read once per process.
var currentInstallChannel = sync.OnceValue(func() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return cliinstallation.InstalledChannel(exe)
})
