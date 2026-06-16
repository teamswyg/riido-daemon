package saasplane

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func newTestPlaneWithToken(t *testing.T, baseURL string, agents []AgentBinding, token string) *Plane {
	t.Helper()
	plane, err := New(Config{
		BaseURL:     baseURL,
		DaemonID:    "daemon-1",
		DeviceID:    "device-1",
		Agents:      agents,
		BearerToken: token,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return plane
}

type fakeAssignmentServer struct {
	t            *testing.T
	server       *httptest.Server
	mu           sync.Mutex
	bearerToken  string
	deviceID     string
	deviceSecret string

	assignmentsByAgent  map[string][]assignmentcontract.Assignment
	assignmentsByID     map[string]assignmentcontract.Assignment
	activeByAgent       map[string]assignmentcontract.Assignment
	cancelByAgent       map[string]assignmentcontract.Assignment
	staleHeartbeatIDs   map[string]bool
	requestCounts       map[string]int
	transientFailures   map[string]int
	transientStatuses   map[string]int
	bindings            []assignmentcontract.AgentRuntimeBinding
	pollRequestsByAgent map[string][]assignmentcontract.PollRequest
	runtimeSnapshots    []DeviceRuntimeSnapshotSyncRequest
	events              []assignmentcontract.AgentEventRequest
	heartbeats          []assignmentcontract.AgentHeartbeatRequest
}

func newFakeAssignmentServer(t *testing.T) *fakeAssignmentServer {
	t.Helper()
	f := &fakeAssignmentServer{
		t:                   t,
		assignmentsByAgent:  map[string][]assignmentcontract.Assignment{},
		assignmentsByID:     map[string]assignmentcontract.Assignment{},
		activeByAgent:       map[string]assignmentcontract.Assignment{},
		cancelByAgent:       map[string]assignmentcontract.Assignment{},
		staleHeartbeatIDs:   map[string]bool{},
		requestCounts:       map[string]int{},
		transientFailures:   map[string]int{},
		transientStatuses:   map[string]int{},
		pollRequestsByAgent: map[string][]assignmentcontract.PollRequest{},
	}
	f.server = httptest.NewServer(http.HandlerFunc(f.handle))
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeAssignmentServer) URL() string {
	return f.server.URL
}

func (f *fakeAssignmentServer) enqueue(assignment assignmentcontract.Assignment) {
	f.assignmentsByAgent[assignment.AgentID] = append(f.assignmentsByAgent[assignment.AgentID], assignment)
	f.assignmentsByID[assignment.ID] = assignment
}

func (f *fakeAssignmentServer) cancelNext(agentID string, assignment assignmentcontract.Assignment) {
	assignment.State = assignmentcontract.AssignmentCancelling
	f.cancelByAgent[agentID] = assignment
	f.assignmentsByID[assignment.ID] = assignment
}

func (f *fakeAssignmentServer) activeNext(agentID string, assignment assignmentcontract.Assignment) {
	if assignment.State == "" {
		assignment.State = assignmentcontract.AssignmentLeased
	}
	f.activeByAgent[agentID] = assignment
	f.assignmentsByID[assignment.ID] = assignment
}

func (f *fakeAssignmentServer) failNext(path string, count, status int) {
	f.transientFailures[path] = count
	f.transientStatuses[path] = status
}

func (f *fakeAssignmentServer) requestCount(path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.requestCounts[path]
}

func (f *fakeAssignmentServer) pollRequestsFor(agentID string) []assignmentcontract.PollRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]assignmentcontract.PollRequest(nil), f.pollRequestsByAgent[agentID]...)
}

func (f *fakeAssignmentServer) handle(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requestCounts[r.URL.Path]++
	if f.deviceSecret != "" && (r.Header.Get("X-Riido-Device-Id") != f.deviceID || r.Header.Get("X-Riido-Device-Secret") != f.deviceSecret) {
		http.Error(w, "missing device credential", http.StatusUnauthorized)
		return
	}
	if f.bearerToken != "" && r.Header.Get("Authorization") != "Bearer "+f.bearerToken {
		http.Error(w, "missing bearer token", http.StatusUnauthorized)
		return
	}
	if f.transientFailures[r.URL.Path] > 0 {
		f.transientFailures[r.URL.Path]--
		status := f.transientStatuses[r.URL.Path]
		if status == 0 {
			status = http.StatusServiceUnavailable
		}
		http.Error(w, "transient failure", status)
		return
	}
	if strings.Trim(r.URL.Path, "/") == "v1/daemon/agent-bindings" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, AgentRuntimeBindingListResponse{
			SchemaVersion: assignmentcontract.SchemaVersion,
			Bindings:      append([]assignmentcontract.AgentRuntimeBinding(nil), f.bindings...),
		})
		return
	}
	if strings.Trim(r.URL.Path, "/") == "v1/daemon/runtime-snapshot" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req DeviceRuntimeSnapshotSyncRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.runtimeSnapshots = append(f.runtimeSnapshots, req)
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, struct {
			SchemaVersion string `json:"schema_version"`
		}{SchemaVersion: assignmentcontract.SchemaVersion})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 4 || parts[0] != "v1" || parts[1] != "agents" {
		http.NotFound(w, r)
		return
	}
	agentID, err := url.PathUnescape(parts[2])
	if err != nil {
		http.Error(w, "bad agent id", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	switch parts[3] {
	case "poll":
		f.handlePoll(w, r, agentID)
	case "heartbeat":
		f.handleHeartbeat(w, r)
	case "events":
		f.handleEvents(w, r, agentID)
	default:
		http.NotFound(w, r)
	}
}
