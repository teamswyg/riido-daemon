package main

import "testing"

func TestDaemonBackgroundStartRejectsExistingDaemon(t *testing.T) {
	paths := newDaemonBackgroundPaths(t)
	if err := run(paths.startArgs()); err != nil {
		t.Fatalf("first background start: %v", err)
	}
	cleanupBackgroundDaemon(t, paths)

	firstPID := readDaemonPIDFile(t, paths.pid)
	err := run(paths.startArgs())
	if err == nil {
		t.Fatal("expected duplicate background start to fail while daemon is running")
	}
	if secondPID := readDaemonPIDFile(t, paths.pid); secondPID != firstPID {
		t.Fatalf("pid file changed after duplicate start: first=%d second=%d", firstPID, secondPID)
	}
}
