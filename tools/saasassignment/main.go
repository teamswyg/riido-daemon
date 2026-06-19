package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

func main() {
	repo := flag.String("repo", ".", "repository root")
	manifest := flag.String("manifest", defaultManifest, "SaaS assignment manifest")
	writeDoc := flag.Bool("write-doc", false, "write generated docs")
	checkDoc := flag.Bool("check-doc", false, "check generated docs")
	evidenceOut := flag.String("evidence-out", "", "write evidence json")
	flag.Parse()

	opts := options{
		Repo: *repo, Manifest: *manifest, WriteDoc: *writeDoc,
		CheckDoc: *checkDoc, EvidenceOut: *evidenceOut,
	}
	if err := run(context.Background(), opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
