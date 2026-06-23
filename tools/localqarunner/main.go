package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	runEvidence := ".riido-local/evidence/local-qa-run.json"
	productEvidence := ".riido-local/evidence/ai-agent-product-acceptance.json"
	releaseEvidence := ".riido-local/evidence/local-release-acceptance.json"
	productLab := ".riido-local/contract-lab/index.html"
	productScreenshots := ".riido-local/screenshots/ai-agent-product-acceptance"
	coverageEvidence := ".riido-local/evidence/local-qa-coverage.json"
	scheduleEvidence := ".riido-local/evidence/local-qa-schedule.json"
	infraEvidence := ".riido-local/evidence/local-qa-dashboard-infra-evidence.json"
	flag.StringVar(&runEvidence, "run-evidence", runEvidence, "local QA run evidence JSON")
	flag.StringVar(&runEvidence, "evidence-out", runEvidence, "alias for -run-evidence")
	cfg := config{
		repo:                 flag.String("repo", ".", "repository root"),
		providerEvidence:     flag.String("provider-evidence", ".riido-local/evidence/provider-real-cli-observation.json", "provider evidence JSON"),
		productEvidence:      flag.String("product-evidence", productEvidence, "product acceptance evidence JSON"),
		releaseEvidence:      flag.String("release-evidence", releaseEvidence, "release install evidence JSON"),
		coverageEvidence:     flag.String("coverage-evidence", coverageEvidence, "local QA coverage snapshot JSON"),
		productLab:           flag.String("product-lab", productLab, "frontend contract lab HTML output"),
		scheduleEvidence:     flag.String("schedule-evidence", scheduleEvidence, "local QA schedule evidence JSON"),
		infraEvidence:        flag.String("infra-evidence", infraEvidence, "private infra dashboard evidence JSON"),
		runEvidence:          &runEvidence,
		dashboardHTML:        flag.String("dashboard", ".riido-local/dashboard/index.html", "dashboard HTML output"),
		coverageManifest:     flag.String("coverage-manifest", "docs/30-architecture/local-acceptance-coverage.riido.json", "coverage manifest JSON"),
		s3Prefix:             flag.String("s3-prefix", os.Getenv("RIIDO_LOCAL_QA_S3_PREFIX"), "optional S3 prefix such as s3://bucket/daily"),
		validFor:             flag.Duration("valid-for", 24*time.Hour, "freshness window for generated evidence"),
		providerTool:         flag.String("provider-tool", "./tools/providerintegrationevidence", "provider evidence tool package"),
		productTool:          flag.String("product-tool", "./tools/localproductacceptance", "product acceptance tool package"),
		releaseTool:          flag.String("release-tool", "./tools/localreleaseacceptance", "release acceptance tool package"),
		dashboardTool:        flag.String("dashboard-tool", "./tools/localqadashboard", "dashboard tool package"),
		clientRoot:           flag.String("client-root", getenvDefault("RIIDO_LOCAL_QA_CLIENT_ROOT", "../riido-client"), "external riido-client worktree"),
		productAgentHost:     flag.String("product-agent-host", getenvDefault("NEXT_PUBLIC_AI_AGENT_HOST", "https://development.ai-api.riido.io"), "AI Agent API host"),
		productRiidoHost:     flag.String("product-riido-api-host", getenvDefault("RIIDO_E2E_RIIDO_API_HOST", "https://development.api.riido.io"), "Riido product API host"),
		productBaseURL:       flag.String("product-base-url", getenvDefault("RIIDO_E2E_BASE_URL", "http://localhost:3000"), "local frontend base URL"),
		productWorkspace:     flag.String("product-workspace-id", os.Getenv("RIIDO_E2E_WORKSPACE_ID"), "workspace id for product contract probes"),
		productTeamID:        flag.String("product-team-id", os.Getenv("RIIDO_E2E_TEAM_ID"), "team id for automatic task fixture creation"),
		productScreenshots:   flag.String("product-screenshots", productScreenshots, "product acceptance screenshot output dir"),
		productStorage:       flag.String("product-storage-state", getenvDefault("RIIDO_E2E_STORAGE_STATE", ".riido-local/private/riido-client-storage-state.json"), "Playwright storage state path"),
		productTaskID:        flag.String("product-task-id", os.Getenv("RIIDO_E2E_TASK_ID"), "task id for product task flow; generated when empty"),
		productAgentID1:      flag.String("product-agent-id-1", os.Getenv("RIIDO_E2E_AGENT_ID_1"), "first agent id for product task mutation flow"),
		productAgentID2:      flag.String("product-agent-id-2", os.Getenv("RIIDO_E2E_AGENT_ID_2"), "second agent id for product task mutation flow"),
		productCommentBody:   flag.String("product-comment-body", os.Getenv("RIIDO_E2E_COMMENT_BODY"), "thread message body for product mutation flow"),
		runIntegration:       flag.Bool("run-integration", true, "run available provider TestIntegration tests"),
		runRelease:           flag.Bool("run-release", true, "run sandboxed daemon install/update acceptance"),
		runProduct:           flag.Bool("run-product", false, "run daemon-owned product acceptance probes"),
		productMutations:     flag.Bool("product-task-mutations", true, "create, verify, and clean up real task assignments in product acceptance"),
		productBrowserE2E:    flag.Bool("product-browser-e2e", false, "deprecated route-browser checks; contract lab is the default"),
		productStartClient:   flag.Bool("product-start-client", false, "start external client dev server when base URL is down"),
		productTaskFixture:   flag.Bool("product-create-task-fixture", true, "create and clean up a development task when product task id is empty"),
		productPrepareDaemon: flag.Bool("product-prepare-saas-daemon", true, "prepare dedicated SaaS-connected local QA daemons for product mutations"),
		continueOnFailure:    flag.Bool("continue-on-failure", true, "continue rendering/upload after provider failures"),
		strictCoverage:       flag.Bool("strict-coverage", false, "fail when coverage_status is not passed"),
	}
	flag.Parse()

	status, err := run(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if status == statusPassed {
		fmt.Println("local-qa-runner: verified")
		return
	}
	fmt.Println("local-qa-runner:", status)
	os.Exit(1)
}
