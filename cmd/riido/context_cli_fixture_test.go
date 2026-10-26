package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

const testRunCapability = "run-cap-should-never-print"

// fakeDaemon answers local IPC requests over net.Pipe and records them.
type fakeDaemon struct {
	mu       sync.Mutex
	requests []map[string]any
	replies  map[string]string
	down     bool
}

func (d *fakeDaemon) dial(_ string, _ time.Duration) (net.Conn, error) {
	if d.down {
		return nil, errors.New("connection refused")
	}
	client, server := net.Pipe()
	go d.serve(server)
	return client, nil
}

func (d *fakeDaemon) serve(conn net.Conn) {
	defer conn.Close()
	var req map[string]any
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		return
	}
	d.mu.Lock()
	d.requests = append(d.requests, req)
	d.mu.Unlock()
	method, _ := req["method"].(string)
	reply, ok := d.replies[method]
	if !ok {
		reply = `{"schema_version":"1","error":"unknown method","method":"` + method + `"}`
	}
	_, _ = conn.Write([]byte(reply + "\n"))
}

func (d *fakeDaemon) lastRequest(t *testing.T) map[string]any {
	t.Helper()
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.requests) == 0 {
		t.Fatal("daemon received no request")
	}
	return d.requests[len(d.requests)-1]
}

type contextHarness struct {
	io     contextIO
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	daemon *fakeDaemon
}

func newContextHarness(t *testing.T, vars map[string]string) *contextHarness {
	t.Helper()
	h := &contextHarness{stdout: &bytes.Buffer{}, stderr: &bytes.Buffer{}, daemon: &fakeDaemon{replies: map[string]string{}}}
	h.io = contextIO{
		stdout:        h.stdout,
		stderr:        h.stderr,
		getenv:        func(k string) string { return vars[k] },
		dial:          h.daemon.dial,
		defaultSocket: func() (string, error) { return "/tmp/test-agentd.sock", nil },
		executable:    func() (string, error) { return "/opt/riido/bin/riido", nil },
		lookPath:      func(string) (string, error) { return "", errors.New("not found") },
		cliVersion:    func() string { return "riido-daemon test" },
	}
	return h
}

func agentVars() map[string]string {
	return map[string]string{
		contextcli.EnvRunToken:     testRunCapability,
		contextcli.EnvTaskID:       "RIID-5839",
		contextcli.EnvWorkspaceID:  "ws-1",
		contextcli.EnvDaemonSocket: "/tmp/run-agentd.sock",
	}
}

// exitCodeOf returns the documented exit status of a command result.
func exitCodeOf(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	var reported reportedExit
	if !errors.As(err, &reported) {
		t.Fatalf("want reportedExit, got %T: %v", err, err)
	}
	return reported.ExitCode()
}

func decodeEnvelope(t *testing.T, out []byte) contextcli.Envelope {
	t.Helper()
	var env contextcli.Envelope
	dec := json.NewDecoder(bytes.NewReader(out))
	if err := dec.Decode(&env); err != nil {
		t.Fatalf("stdout is not one JSON envelope: %v\n%s", err, out)
	}
	if dec.More() {
		t.Fatalf("stdout has more than one JSON value:\n%s", out)
	}
	if env.SchemaVersion != contextcli.EnvelopeSchemaVersion || env.RequestID == "" {
		t.Fatalf("bad envelope header: %+v", env)
	}
	return env
}
