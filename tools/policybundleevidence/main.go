package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	repo := flag.String("repo", ".", "repository root")
	manifest := flag.String("manifest", "docs/20-domain/security/invariants/policy-bundle-loader.riido.json", "policy bundle loader evidence manifest")
	writeDoc := flag.Bool("write-doc", false, "write generated markdown")
	checkDoc := flag.Bool("check-doc", false, "check generated markdown")
	evidenceOut := flag.String("evidence-out", "", "optional evidence JSON output path")
	flag.Parse()

	opts := options{Repo: *repo, Manifest: *manifest, WriteDoc: *writeDoc, CheckDoc: *checkDoc, EvidenceOut: *evidenceOut}
	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
