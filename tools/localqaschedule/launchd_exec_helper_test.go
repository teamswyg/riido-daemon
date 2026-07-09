package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestLocalQALaunchdCommandHelper(t *testing.T) {
	if os.Getenv("RIIDO_LOCALQA_LAUNCHD_HELPER") != "1" {
		return
	}
	args := os.Args
	for i, arg := range args {
		if arg == "--" && i+2 < len(args) {
			os.Stdout.WriteString(args[i+2])
			if args[i+1] != "0" {
				os.Exit(7)
			}
			os.Exit(0)
		}
	}
	os.Exit(8)
}

func fakeLaunchd(t *testing.T, code, output string, calls *[]string) func() {
	t.Helper()
	oldGOOS, oldUID, oldCommand := launchdGOOS, launchdGetuid, launchdCommand
	launchdGOOS = "darwin"
	launchdGetuid = func() int { return 501 }
	launchdCommand = func(name string, args ...string) *exec.Cmd {
		if calls != nil {
			*calls = append(*calls, name+" "+strings.Join(args, " "))
		}
		cmd := exec.Command(os.Args[0], "-test.run=TestLocalQALaunchdCommandHelper", "--", code, output)
		cmd.Env = append(os.Environ(), "RIIDO_LOCALQA_LAUNCHD_HELPER=1")
		return cmd
	}
	return func() { launchdGOOS, launchdGetuid, launchdCommand = oldGOOS, oldUID, oldCommand }
}
