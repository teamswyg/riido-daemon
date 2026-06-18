package main

import (
	"encoding/json"
	"testing"
)

func readDaemonStatus(t *testing.T, sock string) (daemonStatusJSON, string) {
	t.Helper()
	out, err := runCapturingStdout(t, func() error {
		return run([]string{"daemon", "status", "--socket", sock})
	})
	if err != nil {
		t.Fatalf("status: %v\n%s", err, out)
	}
	var status daemonStatusJSON
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		t.Fatalf("parse status %q: %v", out, err)
	}
	return status, out
}
