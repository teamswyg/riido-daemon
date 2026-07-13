package saasplane

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDeviceCredentialUnauthorizedSignalsManagedReload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	plane, err := New(Config{
		BaseURL:      server.URL,
		DaemonID:     "daemon-1",
		DeviceID:     "device-1",
		DeviceSecret: "secret-1",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	var out struct{}
	if err := plane.getJSON(context.Background(), "/v1/daemon/agent-bindings", &out); err == nil {
		t.Fatal("expected unauthorized request error")
	}
	select {
	case err := <-plane.DeviceCredentialRejected():
		if err == nil {
			t.Fatal("expected rejection error")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for device credential rejection")
	}
}

func TestBearerUnauthorizedDoesNotSignalDeviceCredentialReload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	plane, err := New(Config{
		BaseURL:     server.URL,
		DaemonID:    "daemon-1",
		BearerToken: "user-token",
		Agents:      []AgentBinding{{AgentID: "agent-1", RuntimeProvider: "codex"}},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	var out struct{}
	if err := plane.getJSON(context.Background(), "/v1/daemon/agent-bindings", &out); err == nil {
		t.Fatal("expected unauthorized request error")
	}
	select {
	case err := <-plane.DeviceCredentialRejected():
		t.Fatalf("unexpected device credential rejection: %v", err)
	default:
	}
}
