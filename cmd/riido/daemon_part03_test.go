package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestDaemonStartHoldsSingletonLock(t *testing.T) {
	sock := daemonSocketPath(t)
	lockPath := daemonLockPath(t)
	t.Setenv(envTaskQueueDir, "")
	t.Setenv(envTaskReportDir, "")

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		errCh <- runDaemonWithContext(ctx, []string{
			"start", "--foreground",
			"--socket", sock,
			"--lock-file", lockPath,
		})
	}()
	dialDaemon(t, sock, 2*time.Second)

	secondCtx, secondCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer secondCancel()
	err := runDaemonWithContext(secondCtx, []string{
		"start", "--foreground",
		"--socket", daemonSocketPath(t),
		"--lock-file", lockPath,
	})
	if err == nil {
		t.Fatal("expected second daemon start to fail while singleton lock is held")
	}

	cancel()
	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
		t.Fatal("daemon did not exit after context cancel")
	}
}

// TestDaemonHealthEndpoint exercises the health subcommand.
func TestDaemonHealthEndpoint(t *testing.T) {
	sock := daemonSocketPath(t)
	lockPath := daemonLockPath(t)
	t.Setenv(envTaskQueueDir, "")
	t.Setenv(envTaskReportDir, "")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go func() {
		_ = runDaemonWithContext(ctx, []string{"start", "--foreground", "--socket", sock, "--lock-file", lockPath})
	}()
	dialDaemon(t, sock, 2*time.Second)

	out, err := runCapturingStdout(t, func() error {
		return run([]string{"daemon", "health", "--socket", sock})
	})
	if err != nil {
		t.Fatalf("health: %v\n%s", err, out)
	}
	var hr struct {
		Health string `json:"health"`
	}
	if err := json.Unmarshal([]byte(out), &hr); err != nil {
		t.Fatalf("parse: %v\n%s", err, out)
	}
	if hr.Health != "ok" {
		t.Fatalf("health: %q", hr.Health)
	}

	readyOut, err := runCapturingStdout(t, func() error {
		return run([]string{"daemon", "ready", "--socket", sock})
	})
	if err != nil {
		t.Fatalf("ready: %v\n%s", err, readyOut)
	}
	var ready struct {
		Ready             bool   `json:"ready"`
		Readiness         string `json:"readiness"`
		RuntimeCount      int    `json:"runtime_count"`
		RuntimeResponding int    `json:"runtime_responding"`
		SchemaVersion     string `json:"schema_version"`
	}
	if err := json.Unmarshal([]byte(readyOut), &ready); err != nil {
		t.Fatalf("parse ready: %v\n%s", err, readyOut)
	}
	if !ready.Ready || ready.Readiness != "ready" || ready.RuntimeCount != 4 || ready.RuntimeResponding != 4 {
		t.Fatalf("ready payload mismatch: %+v\n%s", ready, readyOut)
	}

	metricsOut, err := runCapturingStdout(t, func() error {
		return run([]string{"daemon", "metrics", "--socket", sock})
	})
	if err != nil {
		t.Fatalf("metrics: %v\n%s", err, metricsOut)
	}
	var metrics struct {
		Metrics struct {
			RuntimeCount        int `json:"runtime_count"`
			RuntimeResponding   int `json:"runtime_responding"`
			ProviderAvailable   int `json:"provider_available"`
			ProviderUnavailable int `json:"provider_unavailable"`
			RunningTasks        int `json:"running_tasks"`
		} `json:"metrics"`
	}
	if err := json.Unmarshal([]byte(metricsOut), &metrics); err != nil {
		t.Fatalf("parse metrics: %v\n%s", err, metricsOut)
	}
	if metrics.Metrics.RuntimeCount != 4 || metrics.Metrics.RuntimeResponding != 4 {
		t.Fatalf("metrics payload mismatch: %+v\n%s", metrics.Metrics, metricsOut)
	}
	if metrics.Metrics.ProviderAvailable+metrics.Metrics.ProviderUnavailable != 4 {
		t.Fatalf("provider metric count mismatch: %+v\n%s", metrics.Metrics, metricsOut)
	}
}

// TestDaemonStopSignalsPidFile spawns a sleep shim, writes its pid to a
// pid file, and verifies `riido daemon stop` sends SIGTERM.
func TestDaemonStopSignalsPidFile(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "agentd.pid")
	shim := exec.Command("/bin/sleep", "30")
	if err := shim.Start(); err != nil {
		t.Fatalf("shim: %v", err)
	}
	waitDone := make(chan error, 1)
	go func() { waitDone <- shim.Wait() }()
	t.Cleanup(func() { _ = shim.Process.Kill() })

	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(shim.Process.Pid)), 0o644); err != nil {
		t.Fatalf("write pid: %v", err)
	}

	if err := run([]string{"daemon", "stop", "--pid-file", pidPath, "--timeout-seconds", "2"}); err != nil {
		t.Fatalf("stop: %v", err)
	}

	select {
	case <-waitDone:
		// process reaped; what matters is that it exited from the signal.
		if shim.ProcessState == nil || shim.ProcessState.Success() {
			// SIGTERM-killed processes report Success()==false. If
			// Success() is true that means /bin/sleep exited normally
			// (which means our stop didn't actually signal it).
			if shim.ProcessState != nil && shim.ProcessState.Success() {
				t.Fatalf("shim exited normally; daemon stop didn't signal it")
			}
		}
	case <-time.After(3 * time.Second):
		t.Fatal("shim did not exit after daemon stop")
	}
}
