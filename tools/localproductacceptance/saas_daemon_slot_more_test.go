package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestPrepareSaaSDaemonSlotEnrollsStartsAndReadsStatus(t *testing.T) {
	binary, logPath := fakeSaaSDaemonBinary(t)
	host := ""
	runDir := t.TempDir()
	workspaceID := "workspace-a"
	cfg := config{
		workspaceID: &workspaceID, agentHost: &host,
		daemonBinary: &binary, daemonRunDir: &runDir,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost ||
			r.URL.Path != "/v2/desktop/workspaces/workspace-a/devices/enroll" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		fmt.Fprint(w, `{"device_id":"dev_slot","device_secret":"sec"}`)
	}))
	defer server.Close()
	host = server.URL

	slot, scenarios := prepareSaaSDaemonSlot(cfg, newAPIClient(server.URL, "tok"), 4)
	if slot.Credential.DeviceID != "dev_slot" || slot.Credential.DeviceSecret != "sec" {
		t.Fatalf("slot credential not populated: %#v", slot.Credential)
	}
	if len(scenarios) != 3 {
		t.Fatalf("expected enroll/start/status scenarios, got %#v", scenarios)
	}
	for _, sc := range scenarios {
		if sc.Status != statusPassed {
			t.Fatalf("scenario failed: %#v", sc)
		}
	}

	logged, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"daemon stop --socket " + slot.Socket,
		"daemon start --socket " + slot.Socket,
		"daemon status --socket " + slot.Socket,
	} {
		if !strings.Contains(string(logged), want) {
			t.Fatalf("missing %q in:\n%s", want, logged)
		}
	}
}
