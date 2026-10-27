package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestRootHelpHidesInternalCommands(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		out := captureHelpStderr(t, func() error { return run([]string{flag}) })
		for _, command := range []string{"daemon", "bridge", "mwsd", "claude-permission-mcp"} {
			if strings.Contains(out, "riido "+command) {
				t.Fatalf("%s exposes internal command %s", flag, command)
			}
		}
		for _, command := range []string{"work", "context", "agents", "meeting", "auth", "whoami", "doctor", "capabilities", "serve", "api", "task", "version"} {
			if !strings.Contains(out, "riido "+command) {
				t.Fatalf("%s lost public command %s", flag, command)
			}
		}
	}
}

func TestDirectDaemonHelpKeepsInternalCommands(t *testing.T) {
	out := captureHelpStderr(t, func() error { return run([]string{"daemon", "--help"}) })
	for _, command := range []string{"daemon start", "daemon stop", "bridge", "mwsd", "task list", "api"} {
		if !strings.Contains(out, "riido "+command) {
			t.Fatalf("direct help lost command %s", command)
		}
	}
}

func captureHelpStderr(t *testing.T, runHelp func() error) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "help")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	previous := os.Stderr
	os.Stderr = file
	defer func() { os.Stderr = previous }()
	if err := runHelp(); err != nil {
		t.Fatal(err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}
