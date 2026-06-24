package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var opt options
	fs := flag.NewFlagSet("localqacandidatedecision", flag.ContinueOnError)
	fs.StringVar(&opt.Repo, "repo", ".", "repository root")
	fs.StringVar(&opt.Manifest, "manifest", defaultManifest, "decision manifest")
	fs.StringVar(&opt.CandidateIn, "candidate-in", "", "local QA run evidence")
	fs.StringVar(&opt.EvidenceOut, "evidence-out", "", "evidence output")
	fs.BoolVar(&opt.WriteDoc, "write-doc", false, "write generated doc")
	fs.BoolVar(&opt.CheckDoc, "check-doc", false, "check generated doc")
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	if err := run(opt); err != nil {
		fmt.Fprintln(os.Stderr, "localqacandidatedecision:", err)
		os.Exit(1)
	}
}
