package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	productEvidence := ".riido-local/evidence/ai-agent-product-acceptance.json"
	coverageEvidence := ".riido-local/evidence/local-qa-coverage.json"
	cfg := config{
		repo:             flag.String("repo", ".", "repository root"),
		s3Prefix:         flag.String("s3-prefix", os.Getenv("RIIDO_LOCAL_QA_S3_PREFIX"), "optional S3 prefix"),
		evidenceOut:      flag.String("evidence-out", ".riido-local/evidence/local-qa-schedule.json", "schedule evidence JSON"),
		productEvidence:  flag.String("product-evidence", productEvidence, "product acceptance evidence JSON"),
		coverageEvidence: flag.String("coverage-evidence", coverageEvidence, "local QA coverage snapshot JSON"),
		clientRoot:       flag.String("client-root", getenvDefault("RIIDO_LOCAL_QA_CLIENT_ROOT", "../riido-client"), "external riido-client worktree"),
		productBaseURL:   flag.String("product-base-url", getenvDefault("RIIDO_E2E_BASE_URL", "http://localhost:3000"), "local frontend base URL"),
		productAgentHost: flag.String("product-agent-host", getenvDefault("NEXT_PUBLIC_AI_AGENT_HOST", "https://staging.ai-api.riido.io"), "AI Agent API host"),
		productRiidoHost: flag.String("product-riido-api-host", getenvDefault("RIIDO_E2E_RIIDO_API_HOST", "https://staging.api.riido.io"), "Riido product API host"),
		productWorkspace: flag.String("product-workspace-id", os.Getenv("RIIDO_E2E_WORKSPACE_ID"), "workspace id for product contract probes"),
		productTeamID:    flag.String("product-team-id", os.Getenv("RIIDO_E2E_TEAM_ID"), "team id for daily task fixture creation"),
		productStorage:   flag.String("product-storage-state", getenvDefault("RIIDO_E2E_STORAGE_STATE", ".riido-local/private/riido-client-storage-state.json"), "Playwright storage state path"),
		productTaskID:    flag.String("product-task-id", os.Getenv("RIIDO_E2E_TASK_ID"), "real task id for daily task mutation flow"),
		productAgentID1:  flag.String("product-agent-id-1", os.Getenv("RIIDO_E2E_AGENT_ID_1"), "first agent id for daily task mutation flow"),
		productAgentID2:  flag.String("product-agent-id-2", os.Getenv("RIIDO_E2E_AGENT_ID_2"), "second agent id for daily task mutation flow"),
		productComment:   flag.String("product-comment-body", os.Getenv("RIIDO_E2E_COMMENT_BODY"), "daily task thread message body"),
		taskMutations:    flag.Bool("product-task-mutations", true, "run daily task mutation probes when a task is configured or discoverable"),
		taskFixture:      flag.Bool("product-create-task-fixture", true, "create and clean up a daily staging task fixture when task id is empty"),
		startClient:      flag.Bool("product-start-client", false, "deprecated route-browser mode only"),
		runProduct:       flag.Bool("run-product", true, "run daemon-owned product acceptance probes"),
		label:            flag.String("label", "io.riido.local-qa", "LaunchAgent label"),
		plistPath:        flag.String("plist", "", "plist path; defaults to ~/Library/LaunchAgents/<label>.plist"),
		hour:             flag.Int("hour", 9, "daily run hour, local time"),
		minute:           flag.Int("minute", 0, "daily run minute, local time"),
		install:          flag.Bool("install", false, "load the LaunchAgent with launchctl"),
		inspect:          flag.Bool("inspect", false, "inspect the loaded LaunchAgent without reinstalling"),
		runAtLoad:        flag.Bool("run-at-load", false, "run once when the LaunchAgent loads"),
	}
	flag.Parse()

	path, err := run(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("local-qa-schedule:", path)
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
