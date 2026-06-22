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
	flag.StringVar(&runEvidence, "run-evidence", runEvidence, "local QA run evidence JSON")
	flag.StringVar(&runEvidence, "evidence-out", runEvidence, "alias for -run-evidence")
	cfg := config{
		repo:              flag.String("repo", ".", "repository root"),
		providerEvidence:  flag.String("provider-evidence", ".riido-local/evidence/provider-real-cli-observation.json", "provider evidence JSON"),
		productEvidence:   flag.String("product-evidence", productEvidence, "product acceptance evidence JSON"),
		runEvidence:       &runEvidence,
		dashboardHTML:     flag.String("dashboard", ".riido-local/dashboard/index.html", "dashboard HTML output"),
		coverageManifest:  flag.String("coverage-manifest", "docs/30-architecture/local-acceptance-coverage.riido.json", "coverage manifest JSON"),
		s3Prefix:          flag.String("s3-prefix", os.Getenv("RIIDO_LOCAL_QA_S3_PREFIX"), "optional S3 prefix such as s3://bucket/daily"),
		validFor:          flag.Duration("valid-for", 24*time.Hour, "freshness window for generated evidence"),
		providerTool:      flag.String("provider-tool", "./tools/providerintegrationevidence", "provider evidence tool package"),
		productTool:       flag.String("product-tool", "./tools/localproductacceptance", "product acceptance tool package"),
		dashboardTool:     flag.String("dashboard-tool", "./tools/localqadashboard", "dashboard tool package"),
		clientRoot:        flag.String("client-root", getenvDefault("RIIDO_LOCAL_QA_CLIENT_ROOT", "../riido-client"), "external riido-client worktree"),
		productBaseURL:    flag.String("product-base-url", getenvDefault("RIIDO_E2E_BASE_URL", "http://localhost:3000"), "local frontend base URL"),
		productWorkspace:  flag.String("product-workspace-id", os.Getenv("RIIDO_E2E_WORKSPACE_ID"), "workspace id for product route probes"),
		runIntegration:    flag.Bool("run-integration", true, "run available provider TestIntegration tests"),
		runProduct:        flag.Bool("run-product", false, "run daemon-owned product acceptance probes"),
		continueOnFailure: flag.Bool("continue-on-failure", true, "continue rendering/upload after provider failures"),
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

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
