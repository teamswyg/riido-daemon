package saasplane

import "net/http"

func (f *fakeAssignmentServer) handle(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requestCounts[r.URL.Path]++
	if !f.allowRequest(w, r) {
		return
	}
	if f.handleTransientFailure(w, r.URL.Path) {
		return
	}
	if f.handleDaemonRoute(w, r) {
		return
	}
	f.handleAgentRoute(w, r)
}
