package main

import (
	"flag"
	"log"
)

func main() {
	var opts options
	flag.StringVar(&opts.Repo, "repo", ".", "repository root")
	flag.StringVar(&opts.Manifest, "manifest", defaultManifest, "manifest path")
	flag.BoolVar(&opts.WriteDoc, "write-doc", false, "rewrite generated doc")
	flag.BoolVar(&opts.CheckDoc, "check-doc", false, "check generated doc")
	flag.StringVar(&opts.EvidenceOut, "evidence-out", "", "write evidence JSON")
	flag.Parse()
	if err := run(opts); err != nil {
		log.Fatal(err)
	}
}
