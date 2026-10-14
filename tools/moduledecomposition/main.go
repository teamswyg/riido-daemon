package main

import (
	"flag"
	"log"
)

func main() {
	opts := options{}
	flag.StringVar(&opts.Repo, "repo", ".", "repository root")
	flag.StringVar(&opts.Manifest, "manifest", defaultManifest, "module decomposition manifest path")
	flag.BoolVar(&opts.WriteDoc, "write-doc", false, "rewrite generated docs")
	flag.BoolVar(&opts.CheckDoc, "check-doc", false, "check generated docs")
	flag.StringVar(&opts.EvidenceOut, "evidence-out", "", "write evidence JSON")
	flag.Parse()
	if err := run(opts); err != nil {
		log.Fatal(err)
	}
}
