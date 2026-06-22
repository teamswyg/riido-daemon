package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	cfg := config{
		repo:            flag.String("repo", ".", "repository root"),
		s3Prefix:        flag.String("s3-prefix", os.Getenv("RIIDO_LOCAL_QA_S3_PREFIX"), "optional S3 prefix"),
		productEvidence: flag.String("product-evidence", os.Getenv("RIIDO_LOCAL_QA_PRODUCT_EVIDENCE"), "optional product acceptance evidence JSON"),
		label:           flag.String("label", "io.riido.local-qa", "LaunchAgent label"),
		plistPath:       flag.String("plist", "", "plist path; defaults to ~/Library/LaunchAgents/<label>.plist"),
		hour:            flag.Int("hour", 9, "daily run hour, local time"),
		minute:          flag.Int("minute", 0, "daily run minute, local time"),
		install:         flag.Bool("install", false, "load the LaunchAgent with launchctl"),
		runAtLoad:       flag.Bool("run-at-load", false, "run once when the LaunchAgent loads"),
	}
	flag.Parse()

	path, err := run(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("local-qa-schedule:", path)
}
