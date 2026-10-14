package main

import (
	"fmt"
	"os"
)

func exitWithRunStatus(status string, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if status == statusPassed {
		fmt.Println("local-qa-runner: verified")
		return
	}
	fmt.Println("local-qa-runner:", status)
	os.Exit(1)
}
