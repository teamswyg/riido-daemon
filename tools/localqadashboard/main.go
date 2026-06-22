package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	input := flag.String("provider-evidence", ".riido-local/evidence/provider-real-cli-observation.json", "provider evidence JSON")
	coverage := flag.String("coverage-manifest", "docs/30-architecture/local-acceptance-coverage.riido.json", "coverage manifest JSON")
	output := flag.String("out", ".riido-local/dashboard/index.html", "dashboard HTML output")
	flag.Parse()

	if err := run(*input, *coverage, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("local-qa-dashboard: rendered")
}
