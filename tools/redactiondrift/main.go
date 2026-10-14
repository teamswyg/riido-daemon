package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	repoRoot := flag.String("repo", ".", "repository root")
	flag.Parse()

	if err := run(*repoRoot); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("redaction-drift: clean")
}
