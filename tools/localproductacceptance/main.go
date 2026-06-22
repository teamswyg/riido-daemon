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
		clientRoot:  flag.String("client-root", clientRoot, "external riido-client worktree to observe read-only"),
		baseURL:     flag.String("base-url", baseURL, "local frontend base URL"),
		workspaceID: flag.String("workspace-id", os.Getenv("RIIDO_E2E_WORKSPACE_ID"), "workspace id for product route probes"),
		evidenceOut: flag.String("evidence-out", out, "product acceptance evidence JSON"),
		validFor:    flag.Duration("valid-for", 24*time.Hour, "freshness window"),
		probeRoutes: flag.Bool("probe-routes", true, "probe local frontend routes"),
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
	fmt.Println("local-product-acceptance:", status)
	os.Exit(1)
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
