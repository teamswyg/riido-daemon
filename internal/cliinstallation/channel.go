package cliinstallation

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// RIID-6036: production is `riido` in $HOME/.riido/bin. The staging test
// channel installs the same package beside it and marks the install directory
// so that one binary picks staging defaults only where the marker exists.
const (
	ChannelMarkerName = "INSTALL_CHANNEL"
	ChannelStaging    = "staging"
)

// InstalledChannel returns ChannelStaging when the directory of the resolved
// executable carries the staging marker, and "" (production) otherwise. A
// missing, unreadable, linked or unknown marker is production.
func InstalledChannel(executable string) string {
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return ""
	}
	return InstalledChannelAt(filepath.Dir(resolved))
}

// InstalledChannelAt reads the marker in an install directory.
func InstalledChannelAt(dir string) string {
	path := filepath.Join(dir, ChannelMarkerName)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 64 {
		return ""
	}
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, 64))
	if err != nil || strings.TrimSpace(string(raw)) != ChannelStaging {
		return ""
	}
	return ChannelStaging
}

// ConfigName is the per-channel folder under the user config or cache
// directory, so staging settings never change production ones.
func ConfigName(channel string) string {
	if channel == ChannelStaging {
		return "riido-cli-staging"
	}
	return "riido-cli"
}

// CommandName is the command a person types for the channel.
func CommandName(channel string) string {
	if channel == ChannelStaging {
		return "riido-staging"
	}
	return "riido"
}
