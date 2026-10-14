package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

func installLaunchAgent(paths schedulePaths) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("launchd install requires macOS")
	}
	domain := "gui/" + strconv.Itoa(os.Getuid())
	_ = exec.Command(paths.launchctl, "bootout", domain, paths.plist).Run()
	cmd := exec.Command(paths.launchctl, "bootstrap", domain, paths.plist)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl bootstrap: %w: %s", err, string(out))
	}
	return nil
}
