package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	kind := flag.String("kind", "", "generator kind")
	specPath := flag.String("spec", "", "spec JSON path")
	templatePath := flag.String("template", "", "Go template path")
	outPath := flag.String("out", "", "output path")
	flag.Parse()

	if err := run(*kind, *specPath, *templatePath, *outPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
