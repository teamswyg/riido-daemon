package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextbroker"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// nativeContextFixture uses the actual host transport and compiled CLI. Only
// runtime activity and the upstream fetch are synthetic; no provider is started.
type nativeContextFixture struct {
	binary, socket, token string
	broker                *contextbroker.Broker
	active, certain       atomic.Bool
	mu                    sync.Mutex
	queries               []contextbroker.Query
	audit                 []string
}

func newNativeContextFixture(t *testing.T) *nativeContextFixture {
	t.Helper()
	f := &nativeContextFixture{}
	f.active.Store(true)
	f.certain.Store(true)
	dir := t.TempDir()
	f.binary = filepath.Join(dir, "riido native.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", f.binary, ".")
	if err := build.Run(); err != nil {
		t.Fatalf("build native CLI: %v", err)
	}
	if runtime.GOOS == "windows" {
		f.socket = fmt.Sprintf(`\\.\pipe\riido-context-test-%d-%d`, os.Getpid(), time.Now().UnixNano())
	} else {
		// macOS limits sockaddr_un to 104 bytes; the Go test temp root can be longer.
		short, err := os.MkdirTemp("", "rctx-")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(short) })
		f.socket = filepath.Join(short, "ipc")
	}
	listener, cleanup, err := listenDaemonSocket(f.socket)
	if err != nil {
		t.Fatal(err)
	}
	f.broker = contextbroker.New(f.socket, filepath.Dir(f.binary))
	f.broker.Fetcher = f
	f.token, err = f.broker.Registry.Issue(contextbroker.Binding{RunKey: "native-run", TaskID: "RIID-5900", WorkspaceID: "native-ws", RuntimeID: "synthetic-runtime", AgentID: "synthetic-agent", AssignmentID: "synthetic-assignment", LeaseToken: "synthetic-lease-private"})
	if err != nil {
		cleanup()
		t.Fatal("issue synthetic capability")
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			f.serve(conn)
		}
	}()
	t.Cleanup(func() {
		cleanup()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("native IPC listener did not stop")
		}
	})
	return f
}

func (f *nativeContextFixture) serve(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	req, ok, err := readDaemonRequest(conn)
	if err != nil || !ok {
		return
	}
	reply := f.broker.Handle(contextcli.ContextReadRequest{Method: string(req.Method), SchemaVersion: req.ContextSchemaVersion, Command: req.Command, Args: req.Args, RunCapability: req.RunCapability, TaskID: req.TaskID, WorkspaceID: req.WorkspaceID, RequestID: req.RequestID}, func(key string) (bool, bool) { return key == "native-run" && f.active.Load(), f.certain.Load() })
	f.mu.Lock()
	f.audit = append(f.audit, contextReadAuditLine(req, reply))
	f.mu.Unlock()
	_ = writeDaemonJSON(conn, reply)
}

func (f *nativeContextFixture) FetchAssignmentContext(_ context.Context, q contextbroker.Query) (contextbroker.Result, error) {
	f.mu.Lock()
	f.queries = append(f.queries, q)
	f.mu.Unlock()
	data, _ := json.Marshal(map[string]any{"marker": "NATIVE_CONTEXT_FIXTURE", "view": q.View, "next_cursor": ""})
	return contextbroker.Result{Data: data, CapturedAt: time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC), Sources: []contextbroker.ResultSource{{Kind: "task", ID: "RIID-5900", Part: "body", Version: "fixture-v1"}}, Live: q.View == "agents" || q.View == "agent"}, nil
}

func (f *nativeContextFixture) run(t *testing.T, override map[string]string, args ...string) (contextcli.Envelope, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, f.binary, append(args, "--json")...)
	vars := map[string]string{contextcli.EnvRunToken: f.token, contextcli.EnvDaemonSocket: f.socket, contextcli.EnvTaskID: "RIID-5900", contextcli.EnvWorkspaceID: "native-ws"}
	maps.Copy(vars, override)
	// A minimal child environment avoids inheriting actual device/provider secrets.
	for _, key := range []string{"PATH", "SystemRoot", "WINDIR", "TEMP", "TMP"} {
		if value := os.Getenv(key); value != "" {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	for key, value := range vars {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatal("native CLI deadline exceeded")
	}
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatalf("native CLI start failed: %T", err)
		}
		code = exit.ExitCode()
	}
	f.assertNoSecrets(t, stdout.String()+stderr.String())
	if stderr.Len() != 0 {
		t.Fatal("JSON CLI wrote unexpected stderr")
	}
	var env contextcli.Envelope
	dec := json.NewDecoder(&stdout)
	if dec.Decode(&env) != nil || env.SchemaVersion != contextcli.EnvelopeSchemaVersion || env.RequestID == "" {
		t.Fatal("native CLI did not return the versioned envelope")
	}
	var extra any
	if dec.Decode(&extra) != io.EOF {
		t.Fatal("native CLI wrote trailing JSON/output")
	}
	return env, code
}

func (f *nativeContextFixture) assertNoSecrets(t *testing.T, output string) {
	t.Helper()
	for _, secret := range []string{f.token, "synthetic-lease-private"} {
		if strings.Contains(output, secret) {
			t.Fatal("credential leaked to CLI output or audit")
		}
	}
}
