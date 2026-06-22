package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	input := flag.String("provider-evidence", ".riido-local/evidence/provider-real-cli-observation.json", "provider evidence JSON")
	external := flag.String("product-evidence", os.Getenv("RIIDO_LOCAL_QA_PRODUCT_EVIDENCE"), "optional product acceptance evidence JSON")
	releaseEvidence := flag.String("release-evidence", ".riido-local/evidence/local-release-acceptance.json", "optional release install evidence JSON")
	runEvidence := flag.String("run-evidence", ".riido-local/evidence/local-qa-run.json", "optional local QA run evidence JSON")
	scheduleEvidence := flag.String("schedule-evidence", ".riido-local/evidence/local-qa-schedule.json", "optional local QA schedule evidence JSON")
	infraEvidence := flag.String("infra-evidence", ".riido-local/evidence/local-qa-dashboard-infra-evidence.json", "optional private infra dashboard evidence JSON")
	coverage := flag.String("coverage-manifest", "docs/30-architecture/local-acceptance-coverage.riido.json", "coverage manifest JSON")
	output := flag.String("out", ".riido-local/dashboard/index.html", "dashboard HTML output")
	flag.Parse()

	if err := run(*input, *external, *releaseEvidence, *runEvidence, *scheduleEvidence, *infraEvidence, *coverage, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("local-qa-dashboard: rendered")
}
