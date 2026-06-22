package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	opts := options{}
	flag.StringVar(&opts.Repo, "repo", ".", "repository root")
	flag.StringVar(&opts.EvidenceOut, "evidence-out", ".riido-local/evidence/local-release-acceptance.json", "release acceptance evidence JSON")
	flag.DurationVar(&opts.ValidFor, "valid-for", 24*time.Hour, "freshness window")
	flag.Parse()

	if err := run(context.Background(), opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("local-release-acceptance: verified")
}
