package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

func inspectLaunchAgent(paths schedulePaths, label string) (launchdEvidence, error) {
	if runtime.GOOS != "darwin" {
		return launchdEvidence{}, fmt.Errorf("launchd inspect requires macOS")
	}
	domain := "gui/" + strconv.Itoa(os.Getuid())
	target := domain + "/" + label
	cmd := exec.Command(paths.launchctl, "print", target)
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
