package main

import (
	"encoding/json"
	"slices"
)

type bridgeProvidersOutput struct {
	SchemaVersion string `json:"schema_version"`
	Providers     []struct {
		Name string `json:"name"`
	} `json:"providers"`
}

func checkBridgeProviders(repo string, manifest Manifest) CheckResult {
	out, err := runGoCommand(repo, "run", "./cmd/riido", "bridge", "providers")
	result := CheckResult{Name: "bridge-providers", Command: "go run ./cmd/riido bridge providers", Pass: err == nil}
	if err != nil {
		result.Detail = err.Error()
		return result
	}
	var decoded bridgeProvidersOutput
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		result.Pass = false
		result.Detail = err.Error()
		return result
	}
	return checkProviderNames(result, decoded, manifest.Providers)
}

func checkProviderNames(result CheckResult, output bridgeProvidersOutput, wants []string) CheckResult {
	names := make([]string, 0, len(output.Providers))
	for _, provider := range output.Providers {
		names = append(names, provider.Name)
	}
	for _, want := range wants {
		if !slices.Contains(names, want) {
			result.Pass = false
			result.Detail = "missing provider: " + want
			return result
		}
	}
	return result
}
