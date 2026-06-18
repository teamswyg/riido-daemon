package main

import "testing"

func TestDaemonStopRequiresSocketOrPID(t *testing.T) {
	err := run([]string{"daemon", "stop"})
	if err == nil {
		t.Fatal("expected error when neither --socket nor --pid-file is provided")
	}
}
