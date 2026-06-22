package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	productEvidence := ".riido-local/evidence/ai-agent-product-acceptance.json"
	cfg := config{
		repo:             flag.String("repo", ".", "repository root"),
		s3Prefix:         flag.String("s3-prefix", os.Getenv("RIIDO_LOCAL_QA_S3_PREFIX"), "optional S3 prefix"),
		evidenceOut:      flag.String("evidence-out", ".riido-local/evidence/local-qa-schedule.json", "schedule evidence JSON"),
		productEvidence:  flag.String("product-evidence", productEvidence, "product acceptance evidence JSON"),
		clientRoot:       flag.String("client-root", getenvDefault("RIIDO_LOCAL_QA_CLIENT_ROOT", "../riido-client"), "external riido-client worktree"),
		productBaseURL:   flag.String("product-base-url", getenvDefault("RIIDO_E2E_BASE_URL", "http://localhost:3000"), "local frontend base URL"),
		productWorkspace: flag.String("product-workspace-id", os.Getenv("RIIDO_E2E_WORKSPACE_ID"), "workspace id for product contract probes"),
		productStorage:   flag.String("product-storage-state", getenvDefault("RIIDO_E2E_STORAGE_STATE", ".riido-local/private/riido-client-storage-state.json"), "Playwright storage state path"),
		startClient:      flag.Bool("product-start-client", false, "deprecated route-browser mode only"),
		runProduct:       flag.Bool("run-product", true, "run daemon-owned product acceptance probes"),
		label:            flag.String("label", "io.riido.local-qa", "LaunchAgent label"),
		plistPath:        flag.String("plist", "", "plist path; defaults to ~/Library/LaunchAgents/<label>.plist"),
		hour:             flag.Int("hour", 9, "daily run hour, local time"),
		minute:           flag.Int("minute", 0, "daily run minute, local time"),
		install:          flag.Bool("install", false, "load the LaunchAgent with launchctl"),
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
