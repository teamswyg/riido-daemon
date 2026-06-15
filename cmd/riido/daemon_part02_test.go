package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// TestDaemonForegroundStartsAndExposesStatus runs the foreground daemon
// in a goroutine, waits until its Unix socket accepts, then checks
// status JSON shape.
func TestDaemonForegroundStartsAndExposesStatus(t *testing.T) {
	sock := daemonSocketPath(t)
	lockPath := daemonLockPath(t)
	t.Setenv(envDaemonID, "daemon-test-1")
	t.Setenv(envDaemonVersion, "riido-agentd v1.2.3")
	t.Setenv(envDaemonProfile, "desktop-api.riido.ai")
	t.Setenv(envServerURL, "https://api.riido.ai")
	t.Setenv(envDeviceName, "MacBook-Pro-SK.local")
	t.Setenv(envRuntimeOwner, "kim")
	t.Setenv(envRuntimeAgents, "Riido, Orion")
	t.Setenv(envWorkspaceCount, "2")
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

	dialDaemon(t, sock, 5*time.Second)

	out, err := runCapturingStdout(t, func() error {
		return run([]string{"daemon", "status", "--socket", sock})
	})
	if err != nil {
		t.Fatalf("status: %v\n%s", err, out)
	}
	var status struct {
		SchemaVersion  string `json:"schema_version"`
		DaemonID       string `json:"daemon_id"`
		DaemonVersion  string `json:"daemon_version"`
		PID            int    `json:"pid"`
		UptimeSeconds  int    `json:"uptime_seconds"`
		Health         string `json:"health"`
		Ready          bool   `json:"ready"`
		Readiness      string `json:"readiness"`
		Profile        string `json:"profile"`
		ServerURL      string `json:"server_url"`
		DeviceName     string `json:"device_name"`
		WorkspaceCount int    `json:"workspace_count"`
		SocketPath     string `json:"socket_path"`
		RunningTasks   int    `json:"running_tasks"`
		Metrics        struct {
			RuntimeCount        int `json:"runtime_count"`
			RuntimeResponding   int `json:"runtime_responding"`
			ProviderAvailable   int `json:"provider_available"`
			ProviderUnavailable int `json:"provider_unavailable"`
			RunningTasks        int `json:"running_tasks"`
		} `json:"metrics"`
		Runtimes []struct {
			RuntimeID  string `json:"runtime_id"`
			Health     string `json:"health"`
			Owner      string `json:"owner"`
			DeviceName string `json:"device_name"`
			Agents     []struct {
				AgentID string `json:"agent_id"`
				Name    string `json:"name"`
				State   string `json:"state"`
			} `json:"agents"`
			Capabilities []struct {
				Provider              string `json:"provider"`
				Available             bool   `json:"available"`
				Reason                string `json:"reason"`
				ProtocolKind          string `json:"protocol_kind"`
				AdapterID             string `json:"adapter_id"`
				AdapterVersion        string `json:"adapter_version"`
				ProtocolVersion       string `json:"protocol_version"`
				CompatibilityStatus   string `json:"compatibility_status"`
				CapabilityFingerprint string `json:"capability_fingerprint"`
			} `json:"capabilities"`
			MaxConcurrent int `json:"max_concurrent"`
		} `json:"runtimes"`
	}
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		t.Fatalf("parse status %q: %v", out, err)
	}
	// Runtimes must expose the provider-specific RuntimeActor pool.
	if len(status.Runtimes) != 4 {
		t.Fatalf("want 4 provider runtimes, got %d: %+v\n%s", len(status.Runtimes), status.Runtimes, out)
	}
	wantProviders := map[string]bool{"claude": false, "codex": false, "openclaw": false, "cursor": false}
	for _, rt := range status.Runtimes {
		if rt.Health != "ok" {
			t.Fatalf("runtime health: %q", rt.Health)
		}
		if rt.Owner != "kim" || rt.DeviceName != "MacBook-Pro-SK.local" {
			t.Fatalf("runtime UI fields missing: owner=%q device=%q\n%s", rt.Owner, rt.DeviceName, out)
		}
		if len(rt.Agents) != 2 || rt.Agents[0].Name != "Riido" || rt.Agents[1].Name != "Orion" {
			t.Fatalf("runtime agents mismatch: %+v\n%s", rt.Agents, out)
		}
		if len(rt.Capabilities) != 1 {
			t.Fatalf("provider runtime should expose exactly one capability, got %d: %+v", len(rt.Capabilities), rt.Capabilities)
		}
		c := rt.Capabilities[0]
		if rt.RuntimeID != "daemon-test-1:"+c.Provider {
			t.Fatalf("runtime_id/provider mismatch: runtime_id=%q provider=%q\n%s", rt.RuntimeID, c.Provider, out)
		}
		if _, ok := wantProviders[c.Provider]; ok {
			wantProviders[c.Provider] = true
		} else {
			t.Fatalf("unexpected provider capability: %+v\n%s", c, out)
		}
		if c.ProtocolKind == "" || c.AdapterID == "" || c.AdapterVersion == "" || c.ProtocolVersion == "" ||
			c.CompatibilityStatus == "" || c.CapabilityFingerprint == "" {
			t.Fatalf("capability missing C3 projection fields: %+v\n%s", c, out)
		}
	}
	for p, seen := range wantProviders {
		if !seen {
			t.Fatalf("capability missing for provider %s", p)
		}
	}
	if status.SchemaVersion == "" {
		t.Fatalf("schema_version missing")
	}
	if status.DaemonID != "daemon-test-1" {
		t.Fatalf("daemon_id mismatch: %q", status.DaemonID)
	}
	if status.DaemonVersion != "riido-agentd v1.2.3" {
		t.Fatalf("daemon_version mismatch: %q", status.DaemonVersion)
	}
	if status.Profile != "desktop-api.riido.ai" || status.ServerURL != "https://api.riido.ai" {
		t.Fatalf("profile/server_url mismatch: profile=%q server=%q", status.Profile, status.ServerURL)
	}
	if status.DeviceName != "MacBook-Pro-SK.local" || status.WorkspaceCount != 2 {
		t.Fatalf("device/workspace mismatch: device=%q workspace=%d", status.DeviceName, status.WorkspaceCount)
	}
	if status.PID == 0 {
		t.Fatalf("pid missing")
	}
	if status.Health != "ok" {
		t.Fatalf("health: %q", status.Health)
	}
	if !status.Ready || status.Readiness != "ready" {
		t.Fatalf("ready: ready=%v readiness=%q", status.Ready, status.Readiness)
	}
	if status.Metrics.RuntimeCount != 4 || status.Metrics.RuntimeResponding != 4 {
		t.Fatalf("runtime metrics mismatch: %+v", status.Metrics)
	}
	if status.Metrics.ProviderAvailable+status.Metrics.ProviderUnavailable != 4 {
		t.Fatalf("provider metrics mismatch: %+v", status.Metrics)
	}
	if status.Metrics.RunningTasks != status.RunningTasks {
		t.Fatalf("running task metrics mismatch: metrics=%d status=%d", status.Metrics.RunningTasks, status.RunningTasks)
	}
	if status.SocketPath != sock {
		t.Fatalf("socket path mismatch: %q vs %q", status.SocketPath, sock)
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("daemon exit error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("daemon did not exit after context cancel")
	}
}
