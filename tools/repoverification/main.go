package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	repo := flag.String("repo", ".", "repository root")
	manifestPath := flag.String("manifest", "docs/readme/verification.riido.json", "verification manifest")
	evidenceOut := flag.String("evidence-out", "", "optional evidence JSON output path")
	writeDoc := flag.Bool("write-doc", false, "write generated markdown")
	checkDoc := flag.Bool("check-doc", false, "check generated markdown")
	runCommands := flag.Bool("run-commands", false, "run verification commands")
	flag.Parse()

	if err := run(*repo, *manifestPath, *evidenceOut, *writeDoc, *checkDoc, *runCommands); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("repo-verification: clean")
}
