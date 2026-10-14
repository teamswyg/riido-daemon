package saasplane

import (
	"net/http"
	"net/url"
	"strings"
)

func (f *fakeAssignmentServer) handleAgentRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 || parts[0] != "v1" || parts[1] != "agents" {
		http.NotFound(w, r)
		return
	}
	agentID, err := url.PathUnescape(parts[2])
	if err != nil {
		http.Error(w, "bad agent id", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	f.dispatchAgentRoute(w, r, agentID, parts[3:])
}

func (f *fakeAssignmentServer) dispatchAgentRoute(
	w http.ResponseWriter,
	r *http.Request,
	agentID string,
	parts []string,
) {
	if parts[0] == "tool-approvals" {
		f.handleToolApprovals(w, r, agentID, parts[1:])
		return
	}
	if len(parts) != 1 {
		http.NotFound(w, r)
		return
	}
	f.dispatchAgentTerminalRoute(w, r, agentID, parts[0])
}

func (f *fakeAssignmentServer) dispatchAgentTerminalRoute(
	w http.ResponseWriter,
	r *http.Request,
	agentID string,
	route string,
) {
	switch route {
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
