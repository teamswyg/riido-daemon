package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var opts options
	flag.StringVar(&opts.Workflow, "workflow", "", "workflow path to verify")
	flag.StringVar(&opts.ID, "id", "", "evidence id")
	flag.StringVar(&opts.Manifest, "manifest", defaultManifest, "CI evidence manifest")
	flag.StringVar(&opts.EvidenceOut, "evidence-out", "", "evidence JSON output path")
	flag.Parse()
	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "cievidence:", err)
		os.Exit(1)
	}
}
