package main

import (
	"fmt"
	"strings"
)

func validateProviderCLINames(names []string) []string {
	if len(names) == 0 {
		return []string{"external_provider_cli_names must not be empty"}
	}
	var problems []string
	for _, name := range names {
		if strings.TrimSpace(name) == "" || strings.ContainsAny(name, `/\`) {
			problems = append(problems, fmt.Sprintf("invalid provider CLI name %q", name))
		}
	}
	return problems
}
