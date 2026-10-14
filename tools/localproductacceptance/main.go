package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	out := ".riido-local/evidence/ai-agent-product-acceptance.json"
	baseURL := getenvDefault("RIIDO_E2E_BASE_URL", "http://localhost:3000")
	clientRoot := getenvDefault("RIIDO_LOCAL_QA_CLIENT_ROOT", "../riido-client")
	cfg := config{
		clientRoot:    flag.String("client-root", clientRoot, "external riido-client worktree to observe read-only"),
		baseURL:       flag.String("base-url", baseURL, "local frontend base URL"),
		apiToken:      flag.String("api-token", firstEnv("RIIDO_AI_AGENT_TOKEN", "RIIDO_AI_AGENT_API_TOKEN"), "development AI Agent API token; never written to evidence"),
		workspaceID:   flag.String("workspace-id", os.Getenv("RIIDO_E2E_WORKSPACE_ID"), "workspace id for contract API probes"),
		taskID:        flag.String("task-id", os.Getenv("RIIDO_E2E_TASK_ID"), "task id for optional assignment flow"),
		firstAgentID:  flag.String("first-agent-id", os.Getenv("RIIDO_E2E_AGENT_ID_1"), "first agent id for optional multi-assignment flow"),
		secondAgentID: flag.String("second-agent-id", os.Getenv("RIIDO_E2E_AGENT_ID_2"), "second agent id for optional multi-assignment flow"),
		evidenceOut:   flag.String("evidence-out", out, "product acceptance evidence JSON"),
		labOut:        flag.String("lab-out", ".riido-local/contract-lab/index.html", "React contract lab HTML output"),
		manualOut:     flag.String("manual-evidence-out", ".riido-local/evidence/manual-qa-evidence.json", "manual QA evidence JSON target exported by the React contract lab"),
		domainCache:   flag.String("domain-cache", getenvDefault("RIIDO_DOMAIN_FIXTURE_CACHE", ".riido-local/evidence/domain-fixture-journey-cache.json"), "domain fixture journey cache JSON"),
		screenshots:   flag.String("screenshots", ".riido-local/screenshots/ai-agent-product-acceptance", "browser screenshot output dir"),
		storageState:  flag.String("storage-state", getenvDefault("RIIDO_E2E_STORAGE_STATE", ".riido-local/private/riido-client-storage-state.json"), "Playwright storage state path"),
		figmaManifest: flag.String("figma-manifest", "docs/30-architecture/figma-ai-agent-daemon-boundary/entries.riido.json", "Figma screen intent manifest"),
		figmaGolden:   flag.String("figma-golden", "docs/30-architecture/figma-ai-agent-daemon-boundary/golden/golden.riido.json", "Figma visual golden manifest"),
		validFor:      flag.Duration("valid-for", 24*time.Hour, "freshness window"),
		probeRoutes:   flag.Bool("probe-routes", true, "probe local frontend routes"),
		browserE2E:    flag.Bool("browser-e2e", false, "run Playwright product browser checks"),
		startClient:   flag.Bool("start-client", false, "start riido-client dev server if base URL is down"),
		agentHost:     flag.String("agent-host", getenvDefault("NEXT_PUBLIC_AI_AGENT_HOST", "https://staging.ai-api.riido.io"), "client AI Agent API host"),
		riidoAPIHost:  flag.String("riido-api-host", getenvDefault("RIIDO_E2E_RIIDO_API_HOST", "https://staging.api.riido.io"), "Riido product API host for task fixtures"),
		teamID:        flag.String("team-id", os.Getenv("RIIDO_E2E_TEAM_ID"), "team id for automatic task fixture creation"),
		taskFixture:   flag.Bool("create-task-fixture", true, "create a staging task fixture when no real task id is available"),
		runMutations:  flag.Bool("run-task-mutations", false, "create real task assignments/comments when ids are supplied"),
		commentBody:   flag.String("comment-body", os.Getenv("RIIDO_E2E_COMMENT_BODY"), "optional task thread message body for mutation flow"),
		prepareDaemon: flag.Bool("prepare-saas-daemon", false, "enroll local QA devices and start dedicated SaaS-connected daemons before task mutations"),
		daemonBinary:  flag.String("daemon-binary", ".riido-local/bin/riido-local-qa", "stable riido binary path used by dedicated local QA daemons"),
		daemonRunDir:  flag.String("daemon-run-dir", ".riido-local/run/product-saas", "runtime directory for dedicated local QA daemon sockets and pid files"),
		daemonSlots:   flag.Int("daemon-slot-count", 2, "number of dedicated local QA SaaS daemons to prepare"),
	}
	flag.Parse()

	status, err := run(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if status == statusPassed {
		fmt.Println("local-product-acceptance: verified")
		return
	}
	if status == statusPartial {
		fmt.Println("local-product-acceptance: partial")
		return
	}
	fmt.Println("local-product-acceptance:", status)
	os.Exit(1)
}
