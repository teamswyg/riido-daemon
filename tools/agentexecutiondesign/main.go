package main

import (
	"flag"
	"log"
)

func main() {
	var opts options
	flag.StringVar(&opts.Repo, "repo", ".", "repository root")
	flag.StringVar(&opts.Manifest, "manifest", "docs/30-architecture/agent-execution-unresolved-design.riido.json", "design manifest")
	flag.BoolVar(&opts.WriteDoc, "write-doc", false, "rewrite generated docs")
	flag.BoolVar(&opts.CheckDoc, "check-doc", false, "verify generated docs")
	flag.StringVar(&opts.EvidenceOut, "evidence-out", "", "write evidence JSON")
	flag.Parse()
	if err := run(opts); err != nil {
		log.Fatal(err)
	}
}
