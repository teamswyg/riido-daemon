package main

import (
	"fmt"
	"strings"
)

func validateContractIdentity(loaded contract) []string {
	var problems []string
	if loaded.SchemaVersion != contractSchemaVersion {
		problems = append(problems, fmt.Sprintf("schema_version must be %q", contractSchemaVersion))
	}
	if strings.TrimSpace(loaded.Product) == "" {
		problems = append(problems, "product is required")
	}
	if loaded.ProviderCLIBundling != "forbidden" {
		problems = append(problems, `provider_cli_bundling must be "forbidden"`)
	}
	return problems
}
