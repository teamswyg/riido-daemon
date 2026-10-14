package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	repo := flag.String("repo", ".", "repository root")
	manifestPath := flag.String("manifest", "docs/30-architecture/provider-real-cli-observation.riido.json", "observation manifest")
	evidenceOut := flag.String("evidence-out", "", "optional evidence JSON output path")
	writeDoc := flag.Bool("write-doc", false, "write generated markdown")
	checkDoc := flag.Bool("check-doc", false, "check generated markdown")
	runIntegration := flag.Bool("run-integration", false, "run available provider TestIntegration tests")
	validFor := flag.Duration("valid-for", 24*time.Hour, "freshness window for generated evidence")
	flag.Parse()

	if err := run(*repo, *manifestPath, *evidenceOut, *writeDoc, *checkDoc, *runIntegration, *validFor); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("provider-integration-evidence: clean")
}
