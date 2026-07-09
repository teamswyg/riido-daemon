package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSaaSDaemonSlotEnvAndStatusSummary(t *testing.T) {
	runDir := t.TempDir()
	cfg := config{daemonRunDir: &runDir}
	cred := saasDeviceCredential{DeviceID: "dev_1", DeviceSecret: "sec", DisplayName: "QA"}
	slot := newSaaSDaemonSlot(cfg, 2, cred)
	if !strings.Contains(slot.Socket, "slot-2/riido.sock") {
		t.Fatalf("unexpected slot paths: %#v", slot)
	}
	env := strings.Join(saasDaemonEnv(slot, "https://staging.ai-api.riido.io"), "\n")
	for _, want := range []string{"RIIDO_DEVICE_ID=dev_1", "RIIDO_DAEMON_PROFILE=staging",
		"RIIDO_WORKDIR_ROOT=" + slot.Workdir} {
		if !strings.Contains(env, want) {
			t.Fatalf("missing env %s in:\n%s", want, env)
		}
	}
	status := summarizeDaemonStatus(map[string]any{
		"ready": true, "daemon_version": "v0.0.60",
		"profile": "staging", "runtimes": []any{"codex", "claude"},
	})
	if status["runtimes_count"] != 2 || status["profile"] != "staging" {
		t.Fatalf("unexpected daemon status summary: %#v", status)
	}
	if saasStatusScenarioID(2) != "local.saas.daemon_status.2" {
		t.Fatal("unexpected status scenario id")
	}
}

func TestSaaSDeviceEnrollScenario(t *testing.T) {
	workspaceID := "ws/1"
	cfg := config{workspaceID: &workspaceID}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/v2/desktop/workspaces/ws%2F1/devices/enroll" {
			t.Fatalf("unexpected enroll path: %s", r.URL.EscapedPath())
		}
		fmt.Fprint(w, `{"device_id":"dev_1","device_secret":"sec",`+
			`"display_name":"Riido Local QA Slot 3"}`)
	}))
	defer server.Close()
	cred, sc := enrollSaaSDevice(cfg, newAPIClient(server.URL, "tok"), 3)
	if sc.Status != statusPassed || cred.DeviceID != "dev_1" || cred.DeviceSecret != "sec" {
		t.Fatalf("unexpected enroll result: cred=%#v scenario=%#v", cred, sc)
	}
	summary := summarizeDeviceEnroll(map[string]any{"device_id": "", "device_secret": "sec"})
	if summary["device_id_present"] != false || summary["device_secret_returned"] != true {
		t.Fatalf("unexpected enroll summary: %#v", summary)
	}
	if saasEnrollScenarioID(3) != "local.saas.device_enroll.3" {
		t.Fatal("unexpected enroll scenario id")
	}
}

func TestEnvHelpersAndNumberFormatting(t *testing.T) {
	t.Setenv("RIIDO_TEST_FIRST", "")
	t.Setenv("RIIDO_TEST_SECOND", "value")
	if getenvDefault("RIIDO_TEST_FIRST", "fallback") != "fallback" {
		t.Fatal("empty env should use fallback")
	}
	if firstEnv("RIIDO_TEST_FIRST", "RIIDO_TEST_SECOND") != "value" {
		t.Fatal("first non-empty env was not selected")
	}
	if intString(42) != "42" || localQAMachineID(4) == localQAMachineID(5) {
		t.Fatal("number formatting or machine id should be stable and slot-specific")
	}
}
