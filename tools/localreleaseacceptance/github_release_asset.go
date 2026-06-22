package main

import "runtime"

func expectedReleaseAsset() string {
	if runtime.GOOS == "windows" {
		return "riido-daemon_windows_" + runtime.GOARCH + ".zip"
	}
	return "riido-daemon_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
}

func hasAsset(release githubRelease, name string) bool {
	for _, asset := range release.Assets {
		if asset.Name == name {
			return true
		}
	}
	return false
}
