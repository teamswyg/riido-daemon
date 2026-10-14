package main

import "testing"

func TestDefaultAgentDaemonPathsForWindows(t *testing.T) {
	userHome := func() (string, error) { return `C:\Users\tester`, nil }
	localAppData := func() (string, error) { return `C:\Users\tester\AppData\Local`, nil }

	socket, err := defaultAgentDaemonSocketForOS("windows", userHome, localAppData)
	if err != nil {
		t.Fatal(err)
	}
	if socket != `\\.\pipe\riido-dev-local-helper-agentd` {
		t.Fatalf("socket = %q", socket)
	}

	workdir, err := defaultAgentDaemonWorkdirRootForOS("windows", userHome, localAppData)
	if err != nil {
		t.Fatal(err)
	}
	if workdir != `C:\Users\tester\AppData\Local\Riido\workspaces` {
		t.Fatalf("workdir = %q", workdir)
	}
}

func TestDefaultAgentDaemonPathsPreserveDarwinDefaults(t *testing.T) {
	userHome := func() (string, error) { return "/Users/tester", nil }
	unusedCache := func() (string, error) {
		t.Fatal("Darwin defaults must not resolve Windows local app data")
		return "", nil
	}

	socket, err := defaultAgentDaemonSocketForOS("darwin", userHome, unusedCache)
	if err != nil {
		t.Fatal(err)
	}
	if socket != "/Users/tester/Library/Application Support/riido/agentd.sock" {
		t.Fatalf("socket = %q", socket)
	}

	workdir, err := defaultAgentDaemonWorkdirRootForOS("darwin", userHome, unusedCache)
	if err != nil {
		t.Fatal(err)
	}
	if workdir != "/Users/tester/Library/Application Support/riido/workspaces" {
		t.Fatalf("workdir = %q", workdir)
	}
}
