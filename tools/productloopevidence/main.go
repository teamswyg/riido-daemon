package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var opts options
	flag.StringVar(&opts.Manifest, "manifest", defaultManifest, "product loop evidence manifest")
	flag.StringVar(&opts.EvidenceOut, "evidence-out", "", "write evidence JSON")
	flag.BoolVar(&opts.WriteDoc, "write-doc", false, "rewrite generated markdown")
	flag.BoolVar(&opts.CheckDoc, "check-doc", false, "check generated markdown")
	flag.BoolVar(&opts.Strict, "strict", false, "fail when product loop evidence remains partial")
	flag.Parse()
	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "productloopevidence:", err)
		os.Exit(1)
	}
}
