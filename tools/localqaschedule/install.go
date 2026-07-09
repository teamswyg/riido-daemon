package main

import (
	"fmt"
	"strconv"
)

func installLaunchAgent(paths schedulePaths) error {
	if launchdGOOS != "darwin" {
		return fmt.Errorf("launchd install requires macOS")
	}
	domain := "gui/" + strconv.Itoa(launchdGetuid())
	_ = launchdCommand(paths.launchctl, "bootout", domain, paths.plist).Run()
	cmd := launchdCommand(paths.launchctl, "bootstrap", domain, paths.plist)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl bootstrap: %w: %s", err, string(out))
	}
	return nil
}
