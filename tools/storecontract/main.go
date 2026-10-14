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
	evidenceOutPath := flag.String("evidence-out", "", "optional JSON evidence output path")
	policyTablePath := flag.String("policy-table", defaultPolicyTablePath, "generated store channel policy table path")
	writePolicyTable := flag.Bool("write-policy-table", false, "rewrite the generated store channel policy table")
	checkPolicyTable := flag.Bool("check-policy-table", false, "verify the generated store channel policy table")
	flag.Parse()
	outputPath, outputErr := selectedOutputPath(*outPath, *evidenceOutPath)
	if outputErr != nil {
		fmt.Fprintln(os.Stderr, outputErr)
		os.Exit(1)
	}

	result, err := runWithOptions(*repoRoot, *contractPath, runOptions{
		PolicyTablePath:  *policyTablePath,
		WritePolicyTable: *writePolicyTable,
		CheckPolicyTable: *checkPolicyTable,
	})
	if outputPath != "" {
		if writeErr := writeJSON(outputPath, result); writeErr != nil {
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
