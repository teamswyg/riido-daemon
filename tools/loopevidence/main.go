package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var opts options
	flag.StringVar(&opts.Repo, "repo", ".", "repository root")
	flag.StringVar(&opts.Manifest, "manifest", defaultManifest, "loop evidence manifest")
	flag.StringVar(&opts.Doc, "doc", "", "generated markdown path; defaults to manifest generated_doc")
	flag.BoolVar(&opts.Write, "write", false, "write generated markdown")
	flag.BoolVar(&opts.Check, "check", false, "check generated markdown is current")
	flag.StringVar(&opts.EvidenceOut, "evidence-out", "", "write evidence JSON")
	flag.Parse()

	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("loop-evidence: clean")
}
