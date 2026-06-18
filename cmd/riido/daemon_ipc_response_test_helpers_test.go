package main

import (
	"encoding/json"
	"testing"
)

func daemonEndpointOutput(t *testing.T, sock string, command daemonCommand) string {
	t.Helper()
	out, err := runCapturingStdout(t, func() error {
		return run([]string{"daemon", string(command), "--socket", sock})
	})
	if err != nil {
		t.Fatalf("%s: %v\n%s", command, err, out)
	}
	return out
}

func decodeDaemonEndpointJSON[T any](t *testing.T, out string) T {
	t.Helper()
	var payload T
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("parse daemon endpoint json: %v\n%s", err, out)
	}
	return payload
}
