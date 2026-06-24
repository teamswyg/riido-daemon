package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var opts options
	flag.StringVar(&opts.Manifest, "manifest", defaultManifest, "loop registry manifest")
	flag.StringVar(&opts.EvidenceOut, "evidence-out", "", "optional evidence JSON output path")
	flag.StringVar(&opts.ChangedFiles, "changed-files", "", "optional newline-delimited changed file list")
	flag.BoolVar(&opts.WriteDoc, "write-doc", false, "write generated markdown")
	flag.BoolVar(&opts.CheckDoc, "check-doc", false, "check generated markdown")
	flag.Parse()
	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "loopregistry:", err)
		os.Exit(1)
	}
	fmt.Println("loopregistry: clean")
}
