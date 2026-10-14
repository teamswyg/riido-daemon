package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	spec := flag.String("spec", "", "source QA DSL JSON")
	out := flag.String("out", "", "generated JSON output")
	flag.Parse()
	if *spec == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "-spec and -out are required")
		os.Exit(2)
	}
	body, err := os.ReadFile(*spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read spec: %v\n", err)
		os.Exit(1)
	}
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		fmt.Fprintf(os.Stderr, "parse spec: %v\n", err)
		os.Exit(1)
	}
	generated, err := json.Marshal(value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode generated JSON: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "create output dir: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*out, append(generated, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write generated JSON: %v\n", err)
		os.Exit(1)
	}
}
