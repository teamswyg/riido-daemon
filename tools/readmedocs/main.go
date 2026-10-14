package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	opts := options{}
	flag.StringVar(&opts.Repo, "repo", ".", "repository root")
	flag.StringVar(&opts.Manifest, "manifest", defaultManifest, "readme pages manifest")
	flag.BoolVar(&opts.WriteDoc, "write-doc", false, "write generated docs")
	flag.BoolVar(&opts.CheckDoc, "check-doc", false, "check generated docs")
	flag.StringVar(&opts.EvidenceOut, "evidence-out", "", "write evidence json")
	flag.Parse()
	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("readme-docs: clean")
}
