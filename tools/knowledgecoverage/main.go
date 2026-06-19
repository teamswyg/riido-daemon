package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	repo := flag.String("repo", ".", "repository root")
	manifestPath := flag.String("manifest", "docs/30-architecture/executable-knowledge.riido.json", "manifest path")
	evidenceOut := flag.String("evidence-out", "", "optional evidence JSON output path")
	writeDoc := flag.Bool("write-doc", false, "write generated reader doc")
	checkDoc := flag.Bool("check-doc", false, "check generated reader doc")
	flag.Parse()

	if err := run(*repo, *manifestPath, *evidenceOut, *writeDoc, *checkDoc); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("knowledge-coverage: clean")
}
