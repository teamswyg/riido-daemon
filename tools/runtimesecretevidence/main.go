package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	repo := flag.String("repo", ".", "repository root")
	manifestPath := flag.String("manifest", "docs/30-architecture/runtime-secret-private-evidence.riido.json", "boundary manifest")
	evidenceOut := flag.String("evidence-out", "", "optional public evidence JSON output path")
	writeDoc := flag.Bool("write-doc", false, "write generated markdown")
	checkDoc := flag.Bool("check-doc", false, "check generated markdown")
	flag.Parse()

	if err := run(*repo, *manifestPath, *evidenceOut, *writeDoc, *checkDoc); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("runtime-secret-evidence: clean")
}
