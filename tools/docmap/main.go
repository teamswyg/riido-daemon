package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	repo := flag.String("repo", ".", "repository root")
	manifestPath := flag.String("manifest", "docs/readme/document-map.riido.json", "document map manifest")
	evidenceOut := flag.String("evidence-out", "", "optional evidence JSON output path")
	write := flag.Bool("write", false, "write generated docs")
	check := flag.Bool("check", false, "check generated docs")
	flag.Parse()

	if err := run(*repo, *manifestPath, *evidenceOut, *write, *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("doc-map: clean")
}
