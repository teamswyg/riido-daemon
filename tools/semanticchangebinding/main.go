package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"
)

func main() {
	var changed string
	opts := options{
		Repo:     ".",
		Manifest: "docs/30-architecture/semantic-change-bindings.riido.json",
	}
	flag.StringVar(&opts.Repo, "repo", opts.Repo, "repository root")
	flag.StringVar(&opts.Manifest, "manifest", opts.Manifest, "semantic binding manifest")
	flag.StringVar(&opts.EvidenceOut, "evidence-out", "", "write evidence JSON")
	flag.StringVar(&changed, "changed-files", "", "comma-separated changed paths")
	flag.Parse()
	if changed != "" {
		opts.ChangedFiles = splitComma(changed)
	}
	if err := run(context.Background(), opts); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func splitComma(s string) []string {
	var out []string
	for part := range strings.SplitSeq(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
