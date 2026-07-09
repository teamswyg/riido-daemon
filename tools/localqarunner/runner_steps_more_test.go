package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunnerStepOrchestrationUsesEvidenceTools(t *testing.T) {
	root := t.TempDir()
	tool := writeRunnerTool(t, root)
	cfg := runnerStepConfig(tool)
	evidence := runEvidence{Status: statusPassed, CoverageStatus: statusPassed}
	runProviderStep(root, cfg, &evidence)
	runReleaseStep(root, cfg, &evidence)
	runProductStep(root, cfg, &evidence)
	runDashboardStep(root, cfg, &evidence)
	runFinalDashboardStep(root, cfg, &evidence)
	want := []string{
		"provider-evidence", "release-acceptance", "product-acceptance",
		"dashboard-render", "dashboard-render-final",
	}
	if len(evidence.Steps) != len(want) {
		t.Fatalf("steps=%+v", evidence.Steps)
	}
	for i, id := range want {
		if evidence.Steps[i].ID != id || evidence.Steps[i].Status != statusPassed {
			t.Fatalf("step[%d]=%+v", i, evidence.Steps[i])
		}
	}
	if len(evidence.ProviderSummary) != 1 || evidence.ProviderSummary[0].ID != "codex" {
		t.Fatalf("provider summary=%+v", evidence.ProviderSummary)
	}
}

func writeRunnerTool(t *testing.T, root string) string {
	t.Helper()
	tool := filepath.Join(root, "tool.go")
	body := `package main
import("flag";"os";"path/filepath";"time")
func main(){
out:=flag.String("evidence-out","","")
for _, n := range []string{"repo","provider-evidence","run-evidence","schedule-evidence","infra-evidence","release-evidence","coverage-manifest","out","coverage-out","product-evidence","client-root","base-url","workspace-id","screenshots","storage-state","agent-host","riido-api-host","team-id","lab-out","manual-evidence-out","domain-cache","task-id","first-agent-id","second-agent-id","comment-body"} { flag.String(n,"","") }
flag.Duration("valid-for",time.Hour,"")
for _, n := range []string{"check-doc","run-integration","browser-e2e","start-client","create-task-fixture","run-task-mutations","prepare-saas-daemon"} { flag.Bool(n,false,"") }
flag.Parse()
if *out != "" { os.MkdirAll(filepath.Dir(*out),0755); os.WriteFile(*out,[]byte("{\"status\":\"passed\",\"providers\":[{\"id\":\"codex\",\"available\":true}]}"),0644) }
}`
	if err := os.WriteFile(tool, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return tool
}

func runnerStepConfig(tool string) config {
	cfg := uploadTestConfig("", "release.json", "coverage.json", "lab.html", "schedule.json", "infra.json")
	cfg.providerTool, cfg.releaseTool, cfg.productTool, cfg.dashboardTool = &tool, &tool, &tool, &tool
	cfg.validFor = strPtrDuration(time.Hour)
	cfg.coverageManifest = strPtr("coverage-manifest.json")
	cfg.clientRoot, cfg.productBaseURL = strPtr("client"), strPtr("http://127.0.0.1")
	cfg.productWorkspace, cfg.productTeamID = strPtr("workspace"), strPtr("team")
	cfg.productScreenshots, cfg.productStorage = strPtr("screens"), strPtr("storage.json")
	cfg.productAgentHost, cfg.productRiidoHost = strPtr("agent"), strPtr("riido")
	cfg.productTaskID, cfg.productAgentID1 = strPtr("task"), strPtr("agent-a")
	cfg.productAgentID2, cfg.productCommentBody = strPtr("agent-b"), strPtr("body")
	yes, no := true, false
	cfg.runIntegration, cfg.productBrowserE2E = &yes, &yes
	cfg.productStartClient, cfg.productTaskFixture = &yes, &no
	cfg.productMutations, cfg.productPrepareDaemon = &yes, &yes
	return cfg
}

func strPtrDuration(value time.Duration) *time.Duration { return &value }
