package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var opts options
	flag.StringVar(&opts.Manifest, "manifest", defaultManifest, "self-improvement manifest")
	flag.StringVar(&opts.EvidenceDir, "evidence-dir", "out", "directory with evidence JSON inputs")
	flag.StringVar(&opts.EvidenceOut, "evidence-out", "", "evidence JSON output path")
	flag.BoolVar(&opts.WriteDoc, "write-doc", false, "write generated markdown")
	flag.BoolVar(&opts.CheckDoc, "check-doc", false, "check generated markdown")
	flag.Parse()
	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "selfimprovementevidence:", err)
		os.Exit(1)
	}
}
