package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	runEvidence := ".riido-local/evidence/local-qa-run.json"
	flag.StringVar(&runEvidence, "run-evidence", runEvidence, "local QA run evidence JSON")
	flag.StringVar(&runEvidence, "evidence-out", runEvidence, "alias for -run-evidence")
	cfg := config{
		repo:              flag.String("repo", ".", "repository root"),
		providerEvidence:  flag.String("provider-evidence", ".riido-local/evidence/provider-real-cli-observation.json", "provider evidence JSON"),
		runEvidence:       &runEvidence,
		dashboardHTML:     flag.String("dashboard", ".riido-local/dashboard/index.html", "dashboard HTML output"),
		coverageManifest:  flag.String("coverage-manifest", "docs/30-architecture/local-acceptance-coverage.riido.json", "coverage manifest JSON"),
		s3Prefix:          flag.String("s3-prefix", os.Getenv("RIIDO_LOCAL_QA_S3_PREFIX"), "optional S3 prefix such as s3://bucket/daily"),
		validFor:          flag.Duration("valid-for", 24*time.Hour, "freshness window for generated evidence"),
		providerTool:      flag.String("provider-tool", "./tools/providerintegrationevidence", "provider evidence tool package"),
		dashboardTool:     flag.String("dashboard-tool", "./tools/localqadashboard", "dashboard tool package"),
		runIntegration:    flag.Bool("run-integration", true, "run available provider TestIntegration tests"),
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
