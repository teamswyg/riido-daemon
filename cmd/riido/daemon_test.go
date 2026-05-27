package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/taskdbplane"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

// daemonSocketPath returns a short Unix socket path. macOS SUN_PATH is
// capped at 104 bytes, so we cannot use t.TempDir() (which on macOS
// produces ~100+ char paths). Instead we allocate a short
// /tmp/riido-<random>.sock and remove it via t.Cleanup.
func daemonSocketPath(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("/tmp", "riido-*.sock")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	path := f.Name()
	_ = f.Close()
	_ = os.Remove(path) // listen will recreate
	t.Cleanup(func() { _ = os.Remove(path) })
	return path
}

func daemonLockPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "agentd.lock")
}

// dialDaemon waits up to deadline for the daemon's local socket to accept.
func dialDaemon(t *testing.T, sock string, deadline time.Duration) {
	t.Helper()
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("daemon socket %s never became reachable", sock)
}

// runCapturingStdout redirects os.Stdout for the duration of fn and
// returns what was written. fn's error is propagated.
func runCapturingStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()
	cmdErr := fn()
	_ = w.Close()
	<-done
	os.Stdout = orig
	return buf.String(), cmdErr
}

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

	ctx, cancel := context.WithCancel(context.Background())
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

func TestDaemonStartHoldsSingletonLock(t *testing.T) {
	sock := daemonSocketPath(t)
	lockPath := daemonLockPath(t)
	t.Setenv(envTaskQueueDir, "")
	t.Setenv(envTaskReportDir, "")

	ctx, cancel := context.WithCancel(context.Background())
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
	ctx, cancel := context.WithCancel(context.Background())
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

// TestDaemonStartUnknownArg rejects unknown flags so typos don't silently
// downgrade the launch.
func TestDaemonStartUnknownArg(t *testing.T) {
	err := run([]string{"daemon", "start", "--bogus"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestLoadDaemonSettingsFromEnv(t *testing.T) {
	env := map[string]string{
		envDaemonID:                      "daemon-1",
		envDaemonVersion:                 "riido-agentd v1.2.3",
		envDaemonProfile:                 "prod",
		envServerURL:                     "https://api.riido.ai",
		envDeviceName:                    "device-a",
		envRuntimeOwner:                  "owner-a",
		envRuntimeAgents:                 "Riido, Orion, ,",
		envWorkspaceCount:                "2",
		envWorkdirRoot:                   "/tmp/riido-workspaces",
		envPolicyBundle:                  "policy-bundle.test.v1",
		envTaskQueueDir:                  "/tmp/riido-queue",
		envTaskReportDir:                 "/tmp/riido-reports",
		envWorkdirRetentionSeconds:       "86400",
		envWorkdirCleanupIntervalSeconds: "300",
		envDaemonPollIntervalSeconds:     "7",
		envDaemonIdlePollIntervalSeconds: "21",
		envDaemonHeartbeatSeconds:        "30",
	}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host-fallback", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.DaemonID != "daemon-1" || settings.DaemonVersion != "riido-agentd v1.2.3" || settings.Profile != "prod" || settings.ServerURL != "https://api.riido.ai" {
		t.Fatalf("daemon fields: %+v", settings)
	}
	if settings.DeviceName != "device-a" || settings.RuntimeOwner != "owner-a" || settings.WorkspaceCount != 2 {
		t.Fatalf("settings mismatch: %+v", settings)
	}
	if settings.WorkdirRoot != "/tmp/riido-workspaces" {
		t.Fatalf("workdir root mismatch: %+v", settings)
	}
	if settings.PolicyBundle != "policy-bundle.test.v1" {
		t.Fatalf("policy bundle mismatch: %+v", settings)
	}
	if settings.PolicyBundleDoc.Version != "policy-bundle.test.v1" ||
		!settings.PolicyBundleDoc.AllowsNativeConfigHook(policy.TrustTierHost, policy.NativeConfigHookClaudeCommandAudit) ||
		!settings.PolicyBundleDoc.AllowsNativeConfigFile(policy.TrustTierHost, policy.NativeConfigFileCodexTaskScopedHome) {
		t.Fatalf("default policy bundle doc mismatch: %+v", settings.PolicyBundleDoc)
	}
	if settings.TaskQueueDir != "/tmp/riido-queue" || settings.TaskReportDir != "/tmp/riido-reports" {
		t.Fatalf("task queue/report dirs mismatch: %+v", settings)
	}
	if settings.WorkdirRetention != 24*time.Hour || settings.WorkdirCleanupEvery != 5*time.Minute {
		t.Fatalf("workdir cleanup settings mismatch: %+v", settings)
	}
	if settings.PollEvery != 7*time.Second || settings.IdlePollEvery != 21*time.Second || settings.HeartbeatEvery != 30*time.Second {
		t.Fatalf("poll/heartbeat settings mismatch: %+v", settings)
	}
	if len(settings.RuntimeAgents) != 2 || settings.RuntimeAgents[0].AgentID != "riido" || settings.RuntimeAgents[1].AgentID != "orion" {
		t.Fatalf("agents: %+v", settings.RuntimeAgents)
	}
}

func TestLoadDaemonSettingsDefaultsTaskReportDirFromQueue(t *testing.T) {
	env := map[string]string{envTaskQueueDir: "/tmp/riido-queue"}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/tmp/riido-queue", "reports")
	if settings.TaskReportDir != want {
		t.Fatalf("task report dir = %q, want %q", settings.TaskReportDir, want)
	}
}

func TestLoadDaemonSettingsRejectsTaskDBSourceWithFileQueue(t *testing.T) {
	env := map[string]string{
		envTaskQueueDir:     "/tmp/riido-queue",
		envTaskDBSourcePath: "/tmp/riido-task-db.json",
	}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected task DB source and file queue conflict")
	}
}

func TestLoadDaemonSettingsAcceptsSaaSControlPlane(t *testing.T) {
	env := map[string]string{
		envSaaSURL:    "https://api.riido.ai",
		envSaaSAgents: "jykim1:codex,jykim2=claude",
		envSaaSToken:  "secret",
	}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.SaaSURL != "https://api.riido.ai" || settings.SaaSAgents != "jykim1:codex,jykim2=claude" || settings.SaaSToken != "secret" {
		t.Fatalf("saas settings = %+v", settings)
	}
}

func TestLoadDaemonSettingsRejectsInvalidPollInterval(t *testing.T) {
	env := map[string]string{envDaemonPollIntervalSeconds: "0"}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected invalid poll interval error")
	}
}

func TestLoadDaemonSettingsRejectsIdlePollBelowActivePoll(t *testing.T) {
	env := map[string]string{
		envDaemonPollIntervalSeconds:     "10",
		envDaemonIdlePollIntervalSeconds: "3",
	}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected idle poll interval below active poll interval error")
	}
}

func TestLoadDaemonSettingsRejectsSaaSWithoutAgents(t *testing.T) {
	env := map[string]string{envSaaSURL: "https://api.riido.ai"}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected SaaS URL without agents error")
	}
}

func TestLoadDaemonSettingsRejectsSaaSWithTaskDBSource(t *testing.T) {
	env := map[string]string{
		envSaaSURL:          "https://api.riido.ai",
		envSaaSAgents:       "jykim1:codex",
		envTaskDBSourcePath: "/tmp/task-db.json",
	}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected SaaS and task DB conflict")
	}
}

func TestLoadDaemonSettingsDefaultWorkdirRoot(t *testing.T) {
	settings, err := loadDaemonSettingsFromEnvWithHome(
		func(string) string { return "" },
		func() (string, error) { return "host", nil },
		func() (string, error) { return "/Users/tester", nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/Users/tester", "Library", "Application Support", "riido", "workspaces")
	if settings.WorkdirRoot != want {
		t.Fatalf("workdir root = %q, want %q", settings.WorkdirRoot, want)
	}
	if settings.PolicyBundle != "policy-bundle.local.v0" {
		t.Fatalf("policy bundle = %q", settings.PolicyBundle)
	}
	if settings.PolicyBundleDoc.Version != policy.DefaultLocalPolicyBundleVersion ||
		!settings.PolicyBundleDoc.AllowsNativeConfigHook(policy.TrustTierHost, policy.NativeConfigHookClaudeCommandAudit) ||
		!settings.PolicyBundleDoc.AllowsNativeConfigFile(policy.TrustTierHost, policy.NativeConfigFileCodexTaskScopedHome) {
		t.Fatalf("default policy bundle doc = %+v", settings.PolicyBundleDoc)
	}
	if settings.DaemonVersion != "riido-agentd v0.0.0" {
		t.Fatalf("daemon version = %q", settings.DaemonVersion)
	}
	if settings.WorkdirRetention != 0 || settings.WorkdirCleanupEvery != 0 {
		t.Fatalf("workdir cleanup should default disabled: %+v", settings)
	}
}

func TestLoadDaemonSettingsLoadsPolicyBundleFile(t *testing.T) {
	path := writePolicyBundleFile(t, "policy-bundle.file.v1")
	env := map[string]string{envPolicyBundlePath: path}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.PolicyBundle != "policy-bundle.file.v1" || settings.PolicyBundlePath != path {
		t.Fatalf("policy bundle settings = %+v", settings)
	}
	if settings.PolicyBundleDoc.Version != "policy-bundle.file.v1" {
		t.Fatalf("policy bundle doc = %+v", settings.PolicyBundleDoc)
	}
}

func TestLoadDaemonSettingsRejectsPolicyBundleVersionMismatch(t *testing.T) {
	path := writePolicyBundleFile(t, "policy-bundle.file.v1")
	env := map[string]string{
		envPolicyBundlePath: path,
		envPolicyBundle:     "policy-bundle.env.v1",
	}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected policy bundle version mismatch error")
	}
}

func TestLoadDaemonSettingsRejectsInvalidPolicyBundleFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy-bundle.riido.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":"wrong"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{envPolicyBundlePath: path}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected invalid policy bundle file error")
	}
}

func TestDaemonToolAutoApproverUsesActivePolicyBundle(t *testing.T) {
	settings := daemonSettings{PolicyBundleDoc: policy.PolicyBundle{
		SchemaVersion:  policy.BundleSchemaVersion,
		Version:        "policy-bundle.tool-auto.v1",
		EffectiveSince: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		TrustTierPolicies: map[policy.TrustTier]policy.TrustTierPolicy{
			policy.TrustTierHost: {
				AllowedSurfaces: policy.AllowedSurfaceSet{
					ToolUse: []policy.ToolUseSurface{policy.ToolUseDestructiveCommand},
				},
			},
		},
	}}
	approver := daemonToolAutoApprover(settings)

	if !approver(agentbridge.ToolRef{Kind: "shell"}) {
		t.Fatal("daemon policy auto approver should approve explicitly allowed shell surface")
	}
	if approver(agentbridge.ToolRef{Kind: "patch_apply"}) {
		t.Fatal("daemon policy auto approver must not approve unallowed patch surface")
	}
}

func TestDaemonToolStartGateUsesActivePolicyBundle(t *testing.T) {
	settings := daemonSettings{PolicyBundleDoc: policy.PolicyBundle{
		SchemaVersion:  policy.BundleSchemaVersion,
		Version:        "policy-bundle.tool-start.v1",
		EffectiveSince: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		TrustTierPolicies: map[policy.TrustTier]policy.TrustTierPolicy{
			policy.TrustTierHost: {
				AllowedSurfaces: policy.AllowedSurfaceSet{
					ToolUse: []policy.ToolUseSurface{policy.ToolUseNetworkEgress},
				},
			},
		},
	}}
	gate := daemonToolStartGate(settings)

	if decision := gate(agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "curl https://example.com"}}); decision.Block {
		t.Fatalf("allowed network surface should not block: %+v", decision)
	}
	if decision := gate(agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "terraform destroy"}}); !decision.Block {
		t.Fatalf("unallowed destructive command should block: %+v", decision)
	}
}

func writePolicyBundleFile(t *testing.T, version string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "policy-bundle.riido.json")
	body := `{
		"schema_version": "riido-policy-bundle.v1",
		"version": "` + version + `",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {}
	}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadDaemonSettingsRejectsInvalidWorkspaceCount(t *testing.T) {
	env := map[string]string{envWorkspaceCount: "nope"}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected invalid workspace count error")
	}
}

func TestLoadDaemonSettingsDefaultsWorkdirCleanupInterval(t *testing.T) {
	env := map[string]string{envWorkdirRetentionSeconds: "7200"}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.WorkdirRetention != 2*time.Hour || settings.WorkdirCleanupEvery != time.Hour {
		t.Fatalf("workdir cleanup settings = %+v", settings)
	}
}

func TestLoadDaemonSettingsRejectsCleanupIntervalWithoutRetention(t *testing.T) {
	env := map[string]string{envWorkdirCleanupIntervalSeconds: "60"}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected cleanup interval without retention error")
	}
}

func TestBuildDaemonControlPlaneUsesMemoryByDefault(t *testing.T) {
	source, reporter, kind, err := buildDaemonControlPlane(daemonSettings{})
	if err != nil {
		t.Fatal(err)
	}
	if kind != "memory" {
		t.Fatalf("kind = %q", kind)
	}
	if _, ok := source.(*controlplane.MemorySource); !ok {
		t.Fatalf("source type = %T", source)
	}
	if _, ok := reporter.(*controlplane.MemoryReporter); !ok {
		t.Fatalf("reporter type = %T", reporter)
	}
}

func TestBuildDaemonControlPlaneUsesSaaS(t *testing.T) {
	source, reporter, kind, err := buildDaemonControlPlane(daemonSettings{
		DaemonID:   "daemon-1",
		DeviceName: "device-1",
		SaaSURL:    "http://127.0.0.1:1",
		SaaSAgents: "jykim1:codex",
		SaaSToken:  "secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	if kind != "saas" {
		t.Fatalf("kind = %q", kind)
	}
	plane, ok := source.(*saasplane.Plane)
	if !ok {
		t.Fatalf("source type = %T", source)
	}
	defer plane.Close()
	if _, ok := reporter.(*saasplane.Plane); !ok {
		t.Fatalf("reporter type = %T", reporter)
	}
}

func TestNewDaemonRuntimeActorsUsesSaaSAgentSlots(t *testing.T) {
	settings := daemonSettings{
		DaemonID:     "daemon-1",
		DeviceName:   "device-1",
		RuntimeOwner: "owner-1",
		SaaSURL:      "https://api.riido.ai",
		SaaSAgents:   "jykim1:codex,jykim2:codex",
		PolicyBundle: "policy-bundle.test.v1",
	}
	runtimes, err := newDaemonRuntimeActors(settings, []agentbridge.Adapter{
		daemonTestAdapter{name: "codex"},
		daemonTestAdapter{name: "claude"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(runtimes) != 2 {
		t.Fatalf("want one runtime per SaaS agent binding, got %d", len(runtimes))
	}
	ctx := context.Background()
	for _, rt := range runtimes {
		if err := rt.Start(ctx); err != nil {
			t.Fatalf("runtime start: %v", err)
		}
		t.Cleanup(func() { _ = rt.Stop(context.Background()) })
	}
	want := map[string]string{
		saasplane.RuntimeIDForAgent("daemon-1", saasplane.AgentBinding{AgentID: "jykim1", RuntimeProvider: "codex"}): "jykim1",
		saasplane.RuntimeIDForAgent("daemon-1", saasplane.AgentBinding{AgentID: "jykim2", RuntimeProvider: "codex"}): "jykim2",
	}
	for _, rt := range runtimes {
		status, err := rt.Status(ctx)
		if err != nil {
			t.Fatalf("status: %v", err)
		}
		agentID, ok := want[status.RuntimeID]
		if !ok {
			t.Fatalf("unexpected runtime id %q", status.RuntimeID)
		}
		if status.MaxConcurrent != 1 {
			t.Fatalf("runtime %s max concurrent = %d", status.RuntimeID, status.MaxConcurrent)
		}
		if len(status.Agents) != 1 || status.Agents[0].AgentID != agentID {
			t.Fatalf("runtime %s agents = %+v", status.RuntimeID, status.Agents)
		}
		if len(status.Capabilities) != 1 || status.Capabilities[0].Provider != "codex" {
			t.Fatalf("runtime %s capabilities = %+v", status.RuntimeID, status.Capabilities)
		}
	}
}

func TestBuildDaemonControlPlaneUsesFileQueue(t *testing.T) {
	queueDir := t.TempDir()
	reportDir := filepath.Join(t.TempDir(), "reports")
	source, reporter, kind, err := buildDaemonControlPlane(daemonSettings{
		TaskQueueDir:  queueDir,
		TaskReportDir: reportDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if kind != "file" {
		t.Fatalf("kind = %q", kind)
	}

	req := bridge.TaskRequest{
		ID:       "task-1",
		Provider: bridge.Provider("claude"),
		Prompt:   "hello",
		Metadata: map[string]string{"workspace_id": "workspace-1"},
	}
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(queueDir, "task-1.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}

	claimed, err := source.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatal(err)
	}
	if claimed == nil || claimed.ID != "task-1" {
		t.Fatalf("claimed = %+v", claimed)
	}
	if err := reporter.StartTask(context.Background(), "task-1"); err != nil {
		t.Fatal(err)
	}
	if err := reporter.CompleteTask(context.Background(), "task-1", agentbridge.Result{Status: agentbridge.ResultCompleted}); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(reportDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one report file, got %d", len(entries))
	}
}

type daemonTestAdapter struct {
	name string
}

func (a daemonTestAdapter) Name() string { return a.name }

func (a daemonTestAdapter) Detect(context.Context, agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{
		Available:         true,
		Executable:        a.name,
		Version:           "test",
		SupportsStreaming: true,
	}, nil
}

func (a daemonTestAdapter) BuildStart(agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return agentbridge.StartCommand{Executable: a.name}, nil
}

func (a daemonTestAdapter) NewParser() agentbridge.Parser { return daemonTestParser{} }

func (a daemonTestAdapter) Translate(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return nil, nil, nil
}

func (a daemonTestAdapter) BlockedArgs() []string { return nil }

type daemonTestParser struct{}

func (daemonTestParser) FeedStdout([]byte) ([]agentbridge.RawEvent, error) { return nil, nil }
func (daemonTestParser) FeedStderr([]byte) ([]agentbridge.RawEvent, error) { return nil, nil }
func (daemonTestParser) Close() ([]agentbridge.RawEvent, error)            { return nil, nil }

func TestBuildDaemonControlPlaneUsesTaskDBSource(t *testing.T) {
	taskDBPath := filepath.Join(t.TempDir(), "task-db.json")
	db := taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{{
			ID:                  "task-1",
			ProjectID:           "workspace-1",
			State:               task.StateQueued,
			Title:               "run from task DB",
			RecommendedProvider: "codex",
		}},
	}
	if err := taskdb.SaveTaskDB(taskDBPath, db); err != nil {
		t.Fatal(err)
	}

	source, reporter, kind, err := buildDaemonControlPlane(daemonSettings{TaskDBSourcePath: taskDBPath})
	if err != nil {
		t.Fatal(err)
	}
	if kind != "taskdb" {
		t.Fatalf("kind = %q", kind)
	}
	if _, ok := source.(*taskdbplane.Plane); !ok {
		t.Fatalf("source type = %T", source)
	}
	if _, ok := reporter.(*taskdbplane.Plane); !ok {
		t.Fatalf("reporter type = %T", reporter)
	}
	claimed, err := source.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatal(err)
	}
	if claimed == nil || claimed.ID != "task-1" || claimed.Provider != "codex" {
		t.Fatalf("claimed = %+v", claimed)
	}
}

func TestBuildDaemonControlPlaneRejectsTaskDBSourceWithReportDir(t *testing.T) {
	_, _, _, err := buildDaemonControlPlane(daemonSettings{
		TaskDBSourcePath: filepath.Join(t.TempDir(), "task-db.json"),
		TaskReportDir:    t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected task DB source and report dir conflict")
	}
}

func TestBuildDaemonControlPlaneRejectsReportDirWithoutQueueDir(t *testing.T) {
	_, _, _, err := buildDaemonControlPlane(daemonSettings{TaskReportDir: t.TempDir()})
	if err == nil {
		t.Fatal("expected error for report dir without queue dir")
	}
}
