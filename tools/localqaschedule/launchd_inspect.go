package main

import (
	"fmt"
	"strconv"
)

func inspectLaunchAgent(paths schedulePaths, label string) (launchdEvidence, error) {
	if launchdGOOS != "darwin" {
		return launchdEvidence{}, fmt.Errorf("launchd inspect requires macOS")
	}
	domain := "gui/" + strconv.Itoa(launchdGetuid())
	target := domain + "/" + label
	cmd := launchdCommand(paths.launchctl, "print", target)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return launchdEvidence{}, fmt.Errorf("launchctl print %s: %w: %s", target, err, string(out))
	}
	live := parseLaunchdPrint(string(out))
	live.Checked = true
	live.Loaded = true
	live.Domain = domain
	return live, nil
}
