package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	repo := flag.String("repo", ".", "repository root")
	manifestPath := flag.String("manifest", "docs/30-architecture/loop-engineering.riido.json", "loop evidence manifest")
	docPath := flag.String("doc", "", "generated markdown path; defaults to manifest generated_doc")
	write := flag.Bool("write", false, "write generated markdown")
	check := flag.Bool("check", false, "check generated markdown is current")
	flag.Parse()

	if err := run(*repo, *manifestPath, *docPath, *write, *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("loop-evidence: clean")
}
