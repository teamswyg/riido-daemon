package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDaemonHelpCommandsHaveNoRuntimeSideEffects(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{name: "daemon", args: []string{"daemon", "--help"}},
		{name: "start", args: []string{"daemon", "start", "--help"}},
		{name: "status", args: []string{"daemon", "status", "--help"}},
		{name: "health", args: []string{"daemon", "health", "--help"}},
		{name: "ready", args: []string{"daemon", "ready", "--help"}},
		{name: "metrics", args: []string{"daemon", "metrics", "--help"}},
		{name: "stop", args: []string{"daemon", "stop", "--help"}},
		{name: "logs", args: []string{"daemon", "logs", "--help"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			spawnCalled := false
			origSpawn := daemonSpawnHelper
			daemonSpawnHelper = func(args []string) (*exec.Cmd, error) {
				spawnCalled = true
				return nil, fmt.Errorf("daemon spawn should not be reached for help")
			}
			t.Cleanup(func() { daemonSpawnHelper = origSpawn })

			if err := run(tc.args); err != nil {
				t.Fatalf("help command returned error: %v", err)
			}
			if spawnCalled {
				t.Fatal("help command reached daemon spawn path")
			}
			lockPath := filepath.Join(home, ".riido", ".lock")
			if _, err := os.Stat(lockPath); err == nil || !os.IsNotExist(err) {
				t.Fatalf("help command touched daemon lock path %s: %v", lockPath, err)
			}
		})
	}
}
