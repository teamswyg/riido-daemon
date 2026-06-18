package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	contractPath := flag.String("contract", "packaging/store/riido_daemon_store_distribution.riido.json", "store distribution contract path")
	repoRoot := flag.String("repo", ".", "repository root")
	outPath := flag.String("out", "", "optional JSON check output path")
	flag.Parse()

	result, err := run(*repoRoot, *contractPath)
	if *outPath != "" {
		if writeErr := writeJSON(*outPath, result); writeErr != nil {
			fmt.Fprintln(os.Stderr, writeErr)
			os.Exit(1)
		}
	}
	if err != nil {
		for _, message := range result.Errors {
			fmt.Fprintln(os.Stderr, message)
		}
		os.Exit(1)
	}
	fmt.Println("store-distribution-contract: clean")
}
