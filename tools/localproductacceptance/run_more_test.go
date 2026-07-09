package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunWritesEvidenceAndLabForIncompleteConfig(t *testing.T) {
	root := t.TempDir()
	cfg := runTestConfig(t, root)
	status, err := run(cfg)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if status != statusFailed {
		t.Fatalf("status=%q, want failed", status)
	}
	data, err := os.ReadFile(*cfg.evidenceOut)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	var evidence evidenceFile
	if err := json.Unmarshal(data, &evidence); err != nil {
		t.Fatalf("decode evidence: %v", err)
	}
	if evidence.SchemaVersion != "riido-product-acceptance.v1" || len(evidence.Scenarios) == 0 {
		t.Fatalf("unexpected evidence: %#v", evidence)
	}
	if findDomainScenario(evidence.Scenarios, "contract.api.bootstrap").Status != statusSkipped {
		t.Fatalf("expected API bootstrap to be skipped without credentials")
	}
	if _, err := os.Stat(*cfg.labOut); err != nil {
		t.Fatalf("contract lab missing: %v", err)
	}
}

func runTestConfig(t *testing.T, root string) config {
	t.Helper()
	clientRoot := filepath.Join(root, "client")
	if err := os.MkdirAll(clientRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	empty := ""
	baseURL := "http://127.0.0.1:3000"
	agentHost := "https://staging.ai-api.riido.io"
	riidoHost := "https://staging.api.riido.io"
	evidenceOut := filepath.Join(root, "evidence.json")
	labOut := filepath.Join(root, "lab", "index.html")
	manualOut := filepath.Join(root, "manual.json")
	domainCache := filepath.Join(root, "domain-cache.json")
	screenshots := filepath.Join(root, "screenshots")
	storageState := filepath.Join(root, "missing-storage.json")
	figmaManifest := filepath.Join(root, "missing-figma.json")
	figmaGolden := filepath.Join(root, "missing-golden.json")
	dur := time.Hour
	disabled := false
	slots := 0
	return config{
		clientRoot: &clientRoot, baseURL: &baseURL, apiToken: &empty,
		workspaceID: &empty, taskID: &empty, firstAgentID: &empty,
		secondAgentID: &empty, evidenceOut: &evidenceOut, labOut: &labOut,
		manualOut: &manualOut, domainCache: &domainCache, screenshots: &screenshots,
		storageState: &storageState, figmaManifest: &figmaManifest, figmaGolden: &figmaGolden,
		validFor: &dur, probeRoutes: &disabled, browserE2E: &disabled,
		startClient: &disabled, agentHost: &agentHost, riidoAPIHost: &riidoHost,
		teamID: &empty, taskFixture: &disabled, runMutations: &disabled,
		commentBody: &empty, prepareDaemon: &disabled, daemonBinary: &empty,
		daemonRunDir: &root, daemonSlots: &slots,
	}
}
