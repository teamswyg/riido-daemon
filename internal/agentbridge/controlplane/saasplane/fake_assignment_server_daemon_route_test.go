package saasplane

import (
	"encoding/json"
	"net/http"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func (f *fakeAssignmentServer) handleDaemonRoute(w http.ResponseWriter, r *http.Request) bool {
	switch strings.Trim(r.URL.Path, "/") {
	case "v1/daemon/agent-bindings":
		f.handleAgentBindings(w, r)
	case "v1/daemon/runtime-snapshot":
		f.handleRuntimeSnapshot(w, r)
	default:
		return false
	}
	return true
}

func (f *fakeAssignmentServer) handleAgentBindings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, AgentRuntimeBindingListResponse{
		SchemaVersion: assignmentcontract.SchemaVersion,
		Bindings:      append([]assignmentcontract.AgentRuntimeBinding(nil), f.bindings...),
	})
}

func (f *fakeAssignmentServer) handleRuntimeSnapshot(w http.ResponseWriter, r *http.Request) {
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
}
