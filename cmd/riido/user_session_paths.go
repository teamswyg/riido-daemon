package main

import (
	"path/filepath"

	"github.com/teamswyg/riido-daemon/internal/cliinstallation"
	"github.com/teamswyg/riido-daemon/internal/securecredential"
)

func userSessionPathsForOS(goos, executable, configDir, cacheDir string) (string, string, error) {
	if !filepath.IsAbs(executable) {
		return "", "", securecredential.ErrUnavailable
	}
	// A staging-channel install keeps its locks apart from production (RIID-6036).
	name := cliinstallation.ConfigName(cliinstallation.InstalledChannelAt(filepath.Dir(executable)))
	switch goos {
	case "darwin":
		if filepath.IsAbs(configDir) {
			return filepath.Join(filepath.Dir(executable), "RiidoCLIAuthHelper.app", "Contents", "MacOS", "riido-cli-auth-helper"), filepath.Join(configDir, name, "auth-locks"), nil
		}
	case "windows":
		if filepath.IsAbs(cacheDir) {
			return filepath.Join(filepath.Dir(executable), "riido-cli-auth-helper.exe"), filepath.Join(cacheDir, name, "auth-locks"), nil
		}
	}
	return "", "", securecredential.ErrUnavailable
}
