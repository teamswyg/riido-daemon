package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	out := flag.String("out", "cmd/riido/provider_model_catalog.generated.json", "output path")
	flag.Parse()
	catalog, err := buildCatalog()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	body, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(*out, append(body, '\n'), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildCatalog() (catalog, error) {
	providers := make(map[string][]model)
	var err error
	if providers["cursor"], err = cursorModels(); err != nil {
		return catalog{}, fmt.Errorf("cursor model catalog: %w", err)
	}
	if providers["claude"], err = claudeModels(); err != nil {
		return catalog{}, fmt.Errorf("claude model catalog: %w", err)
	}
	return catalog{
		SchemaVersion: "riido-provider-model-catalog.v1",
		Providers:     providers,
	}, nil
}
