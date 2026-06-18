package saasplane

import "net/http"

func (f *fakeAssignmentServer) allowRequest(w http.ResponseWriter, r *http.Request) bool {
	if f.deviceSecret != "" && !f.hasDeviceCredential(r) {
		http.Error(w, "missing device credential", http.StatusUnauthorized)
		return false
	}
	if f.bearerToken != "" && r.Header.Get("Authorization") != "Bearer "+f.bearerToken {
		http.Error(w, "missing bearer token", http.StatusUnauthorized)
		return false
	}
	return true
}

func (f *fakeAssignmentServer) hasDeviceCredential(r *http.Request) bool {
	return r.Header.Get("X-Riido-Device-Id") == f.deviceID &&
		r.Header.Get("X-Riido-Device-Secret") == f.deviceSecret
}

func (f *fakeAssignmentServer) handleTransientFailure(w http.ResponseWriter, path string) bool {
	if f.transientFailures[path] <= 0 {
		return false
	}
	f.transientFailures[path]--
	status := f.transientStatuses[path]
	if status == 0 {
		status = http.StatusServiceUnavailable
	}
	http.Error(w, "transient failure", status)
	return true
}
