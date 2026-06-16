package main

import (
	"encoding/json"
	"io"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func TestDaemonShutdownRequestCarriesForcedLevel(t *testing.T) {
	server, client := net.Pipe()
	shutdownCh := make(chan lifecycle.ShutdownLevel, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		handleDaemonConn(server, startFlags{}, daemonSettings{}, time.Now(), nil, shutdownCh, logging.NewWriterLogger(io.Discard))
	}()
	t.Cleanup(func() { _ = client.Close() })
	_ = client.SetDeadline(time.Now().Add(time.Second))

	if err := json.NewEncoder(client).Encode(daemonRequest{Method: daemonMethodShutdown, ShutdownLevel: "forced"}); err != nil {
		t.Fatalf("encode shutdown request: %v", err)
	}
	var ack map[string]string
	if err := json.NewDecoder(client).Decode(&ack); err != nil {
		t.Fatalf("decode shutdown ack: %v", err)
	}
	if ack["shutdown"] != "accepted" || ack["shutdown_level"] != lifecycle.ShutdownForced.String() {
		t.Fatalf("shutdown ack = %+v", ack)
	}
	select {
	case level := <-shutdownCh:
		if level != lifecycle.ShutdownForced {
			t.Fatalf("shutdown level = %s, want %s", level, lifecycle.ShutdownForced)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown level was not delivered")
	}
	<-done
}

func TestTryShutdownViaSocketSendsForcedLevel(t *testing.T) {
	sock := daemonSocketPath(t)
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	received := make(chan daemonRequest, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		var req daemonRequest
		_ = json.NewDecoder(conn).Decode(&req)
		received <- req
		writeShutdownAck(conn, req.lifecycleShutdownLevel())
		_ = ln.Close()
	}()

	if ok := tryShutdownViaSocket(sock, time.Second, lifecycle.ShutdownForced); !ok {
		t.Fatal("forced shutdown socket request did not complete")
	}
	select {
	case req := <-received:
		if req.Method != daemonMethodShutdown || !req.Force || req.ShutdownLevel != lifecycle.ShutdownForced.String() {
			t.Fatalf("shutdown request = %+v", req)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown request was not received")
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
		settings.PolicyBundleDoc.AllowsNativeConfigFile(policy.TrustTierHost, policy.NativeConfigFileCodexTaskScopedHome) {
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
