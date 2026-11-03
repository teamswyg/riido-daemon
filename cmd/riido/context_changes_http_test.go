package main

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
	"github.com/teamswyg/riido-daemon/internal/contextbroker"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

const httpChangesData = `{"snapshot_id":"asg-1","since":"previous","changes":[],"complete":true,"full_refresh_required":false}`

func TestContextChangesCLIThroughIPCAndHTTP(t *testing.T) {
	for _, test := range []struct {
		name         string
		status, exit int
		message      string
	}{
		{"complete", 200, 0, ""},
		{"partial", 200, 0, ""},
		{"budget", 200, 9, ""},
		{"baseline", 410, 5, "expired"},
		{"lease", 403, 4, "invalid lease"},
		{"ended", 409, 10, "ended"},
		{"quota", 429, 7, "quota"},
		{"old server", 400, 8, `json: unknown field "since"`},
		{"live", 200, 6, ""},
		{"wrong outer", 200, 6, ""},
		{"wrong inner", 200, 6, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				var req saasplane.AssignmentContextRequest
				if json.NewDecoder(r.Body).Decode(&req) != nil || req.AssignmentID != "asg-1" || req.LeaseToken != "synthetic-lease" || req.Since != "previous" || req.View != "changes" || req.Target != "" || req.Cursor != "" || r.Method != http.MethodPost || r.URL.Path != "/v1/agents/agent-1/assignment-context" || r.Header.Get("X-Riido-Device-Secret") != "synthetic-device" {
					t.Error("incorrect current binding request")
				}
				w.Header().Set("Content-Type", "application/json")
				if test.status != 200 {
					w.WriteHeader(test.status)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": test.message})
					return
				}
				data := httpChangesData
				if test.name == "partial" {
					data = strings.Replace(strings.Replace(data, `"complete":true`, `"complete":false`, 1), `"full_refresh_required":false`, `"full_refresh_required":true`, 1)
				}
				if test.name == "wrong inner" {
					data = strings.Replace(data, "previous", "other", 1)
				}
				out := saasplane.AssignmentContextResponse{AssignmentID: "asg-1", TaskID: "RIID-5839", View: "changes", CapturedAt: time.Now(), Data: json.RawMessage(data), Live: test.name == "live"}
				if test.name == "wrong outer" {
					out.AssignmentID = "other"
				}
				if err := json.NewEncoder(w).Encode(out); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			plane, err := saasplane.New(saasplane.Config{BaseURL: server.URL, DaemonID: "daemon-1", DeviceID: "device-1", DeviceSecret: "synthetic-device"})
			if err != nil {
				t.Fatal(err)
			}
			h := changesHTTPHarness(t, plane)
			args := []string{"changes", "--since", "previous", "--json", "--no-input"}
			if test.name == "budget" {
				args = append(args, "--budget", "1")
			}
			if got := exitCodeOf(t, runContext(h.io, args)); got != test.exit {
				t.Fatalf("exit=%d, want %d", got, test.exit)
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			if calls.Load() != 1 || strings.Contains(h.stdout.String()+h.stderr.String(), "synthetic-") {
				t.Fatal("duplicate read or credential leak")
			}
			if test.exit != 0 && (env.Error == nil || env.Data != nil) {
				t.Fatal("failed comparison exposed data")
			}
			data, _ := json.Marshal(env.Data)
			if test.name == "partial" && !strings.Contains(string(data), `"full_refresh_required":true`) {
				t.Fatal("partial result lost")
			}
		})
	}
}

func changesHTTPHarness(t *testing.T, plane *saasplane.Plane) *contextHarness {
	t.Helper()
	b := contextbroker.New("", "")
	b.Fetcher = saasContextFetcher{plane: plane}
	grant, err := b.Registry.Issue(contextbroker.Binding{RunKey: "run-1", TaskID: "RIID-5839", WorkspaceID: "ws-1", AgentID: "agent-1", AssignmentID: "asg-1", LeaseToken: "synthetic-lease"})
	if err != nil {
		t.Fatal(err)
	}
	vars := agentVars()
	vars[contextcli.EnvRunToken] = grant
	h := newContextHarness(t, vars)
	h.io.dial = func(_ string, _ time.Duration) (net.Conn, error) {
		client, server := net.Pipe()
		go func() {
			defer server.Close()
			var req contextcli.ContextReadRequest
			if json.NewDecoder(server).Decode(&req) != nil {
				return
			}
			_ = json.NewEncoder(server).Encode(b.Handle(req, func(key string) (bool, bool) { return key == "run-1", true }))
		}()
		return client, nil
	}
	return h
}
