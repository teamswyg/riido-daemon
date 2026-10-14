package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("RIIDO_DAEMON_HELPER") == "1" {
		os.Exit(runDaemonHelperProcess())
	}
	daemonSpawnHelper = spawnDaemonTestBinary
	os.Exit(m.Run())
}

func runDaemonHelperProcess() int {
	args := daemonHelperArgs(os.Args)
	if err := run(args); err != nil {
		_, _ = os.Stderr.WriteString("helper: " + err.Error() + "\n")
		return 1
	}
	return 0
}

func daemonHelperArgs(args []string) []string {
	for i, arg := range args {
		if arg == "--" {
			return args[i+1:]
		}
	}
	return nil
}

func spawnDaemonTestBinary(args []string) (*exec.Cmd, error) {
	spec := []string{"-test.run=TestNeverMatchesAnything"}
	spec = append(spec, "--")
	spec = append(spec, args...)
	cmd := exec.Command(os.Args[0], spec...)
	cmd.Env = append(os.Environ(), "RIIDO_DAEMON_HELPER=1")
	return cmd, nil
}

func TestNeverMatchesAnything(t *testing.T) {}
