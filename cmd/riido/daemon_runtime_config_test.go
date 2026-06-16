package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
)

func TestBuildDaemonControlPlaneUsesSaaS(t *testing.T) {
	source, reporter, kind, err := buildDaemonControlPlane(daemonSettings{
		DaemonID:     "daemon-1",
		DeviceName:   "device-1",
		SaaSURL:      "http://127.0.0.1:1",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
	}, time.Time{})
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

func TestBuildDaemonControlPlaneSaaSUsesDefaultLongPollWait(t *testing.T) {
	pollSeen := make(chan assignmentcontract.PollRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/daemon/agent-bindings":
			writeDaemonRuntimeTestJSON(t, w, saasplane.AgentRuntimeBindingListResponse{
				SchemaVersion: assignmentcontract.SchemaVersion,
				Bindings: []assignmentcontract.AgentRuntimeBinding{{
					AgentID:         "agent-long",
					DaemonID:        "device-1",
					DeviceID:        "device-1",
					RuntimeID:       "device-1:codex",
					RuntimeProvider: "codex",
				}},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/agents/agent-long/poll":
			var req assignmentcontract.PollRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("poll request decode: %v", err)
			}
			pollSeen <- req
			writeDaemonRuntimeTestJSON(t, w, assignmentcontract.PollResponse{
				SchemaVersion: assignmentcontract.SchemaVersion,
				Action:        assignmentcontract.PollNone,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	source, reporter, kind, err := buildDaemonControlPlane(daemonSettings{
		DaemonID:     "device-1",
		DeviceName:   "device-1",
		SaaSURL:      server.URL,
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
	}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closer, ok := source.(interface{ Close() }); ok {
			closer.Close()
		}
	})
	sourcePlane, sourceOK := source.(*saasplane.Plane)
	reporterPlane, reporterOK := reporter.(*saasplane.Plane)
	if kind != "saas" || !sourceOK || !reporterOK || sourcePlane != reporterPlane {
		t.Fatalf("control plane kind/source/reporter = %q %T %T", kind, source, reporter)
	}
	req, err := source.ClaimTask(context.Background(), "device-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if req != nil {
		t.Fatalf("empty fake server should not claim task: %+v", req)
	}
	select {
	case poll := <-pollSeen:
		if poll.WaitMs != 30000 {
			t.Fatalf("wait_ms = %d, want 30000", poll.WaitMs)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for poll request")
	}
}

func TestNewDaemonRuntimeActorsUsesProviderSlotsForDynamicSaaSBindings(t *testing.T) {
	settings := daemonSettings{
		DaemonID:     "daemon-1",
		DeviceName:   "device-1",
		RuntimeOwner: "owner-1",
		SaaSURL:      "https://api.riido.ai",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
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
		t.Fatalf("want one runtime per provider adapter, got %d", len(runtimes))
	}
	ctx := context.Background()
	want := map[string]string{
		"daemon-1:codex":  "codex",
		"daemon-1:claude": "claude",
	}
	for _, rt := range runtimes {
		if err := rt.Start(ctx); err != nil {
			t.Fatalf("runtime start: %v", err)
		}
		t.Cleanup(func() { _ = rt.Stop(context.Background()) })
		status, err := rt.Status(ctx)
		if err != nil {
			t.Fatalf("status: %v", err)
		}
		provider, ok := want[status.RuntimeID]
		if !ok {
			t.Fatalf("unexpected runtime id %q", status.RuntimeID)
		}
		if len(status.Agents) != 0 {
			t.Fatalf("dynamic runtime %s should not use static agents: %+v", status.RuntimeID, status.Agents)
		}
		if len(status.Capabilities) != 1 || status.Capabilities[0].Provider != provider {
			t.Fatalf("runtime %s capabilities = %+v", status.RuntimeID, status.Capabilities)
		}
	}
}

func writeDaemonRuntimeTestJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("write json: %v", err)
	}
}

func TestBuildDaemonControlPlaneUsesFileQueue(t *testing.T) {
	queueDir := t.TempDir()
	reportDir := filepath.Join(t.TempDir(), "reports")
	source, reporter, kind, err := buildDaemonControlPlane(daemonSettings{
		TaskQueueDir:  queueDir,
		TaskReportDir: reportDir,
	}, time.Time{})
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

func TestCodexRuntimeModelsReadConfiguredDefaultModel(t *testing.T) {
	home := t.TempDir()
	configDir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	config := []byte("model = \"gpt-5.5\"\nmodel_reasoning_effort = \"xhigh\"\n")
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), config, 0o600); err != nil {
		t.Fatal(err)
	}

	models := codexRuntimeModels(func() (string, error) { return home, nil })
	if len(models) != 1 || models[0].ModelID != "gpt-5.5" || models[0].Label != "gpt-5.5" || !models[0].IsDefault {
		t.Fatalf("models = %+v", models)
	}
}

func TestCodexRuntimeModelsMissingConfigDoesNotInventModel(t *testing.T) {
	models := codexRuntimeModels(func() (string, error) { return t.TempDir(), nil })
	if len(models) != 0 {
		t.Fatalf("models = %+v", models)
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
