package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var opts options
	flag.StringVar(&opts.Repo, "repo", ".", "repository root")
	flag.StringVar(&opts.Manifest, "manifest", defaultManifest, "provider runtime responsibility manifest")
	flag.BoolVar(&opts.WriteDoc, "write-doc", false, "rewrite generated docs")
	flag.BoolVar(&opts.CheckDoc, "check-doc", false, "check generated docs")
	flag.StringVar(&opts.EvidenceOut, "evidence-out", "", "write evidence JSON")
	flag.Parse()
	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("provider-runtime-responsibility-docs: clean")
}
