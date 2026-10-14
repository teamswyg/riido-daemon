package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

func main() {
	repo := flag.String("repo", ".", "repository root")
	manifest := flag.String("manifest", defaultManifest, "branch gate manifest")
	writeDoc := flag.Bool("write-doc", false, "write generated reader doc")
	checkDoc := flag.Bool("check-doc", false, "check generated reader doc")
	writeScript := flag.Bool("write-script", false, "write generated script")
	checkScript := flag.Bool("check-script", false, "check generated script")
	evidenceOut := flag.String("evidence-out", "", "write evidence json")
	flag.Parse()

	opts := options{Repo: *repo, Manifest: *manifest, WriteDoc: *writeDoc, CheckDoc: *checkDoc}
	opts.WriteScript, opts.CheckScript, opts.EvidenceOut = *writeScript, *checkScript, *evidenceOut
	if err := run(context.Background(), opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
