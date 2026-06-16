package main

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// BridgeProvidersSchemaVersion identifies the JSON shape printed by
// `riido bridge providers`.
const BridgeProvidersSchemaVersion = "riido-bridge-providers.v1"

// BridgeDetectSchemaVersion identifies the JSON shape printed by
// `riido bridge detect`.
const BridgeDetectSchemaVersion = "riido-bridge-detect.v1"

// providerEntry is the JSON-printable view of a registered adapter.
type providerEntry struct {
	Name              string                    `json:"name"`
	BlockedArgs       []string                  `json:"blocked_args"`
	DefaultExecutable string                    `json:"default_executable"`
	Detect            *agentbridge.DetectResult `json:"detect,omitempty"`
}

func runBridge(args []string) error {
	if len(args) < 1 {
		printUsage()
		return fmt.Errorf("missing bridge subcommand")
	}
	switch args[0] {
	case "providers":
		return runBridgeProviders(args[1:])
	case "detect":
		return runBridgeDetect(args[1:])
	default:
		printUsage()
		return fmt.Errorf("unknown bridge subcommand: %s", args[0])
	}
}

func runBridgeProviders(_ []string) error {
	adapters := builtinAgentAdapters()
	entries := make([]providerEntry, 0, len(adapters))
	for _, adapter := range adapters {
		entries = append(entries, providerEntryForAdapter(adapter, nil))
	}
	return printJSON(struct {
		SchemaVersion string          `json:"schema_version"`
		Providers     []providerEntry `json:"providers"`
	}{
		SchemaVersion: BridgeProvidersSchemaVersion,
		Providers:     entries,
	})
}

func runBridgeDetect(_ []string) error {
	ctx := context.Background()
	adapters := builtinAgentAdapters()
	entries := make([]providerEntry, 0, len(adapters))
	for _, adapter := range adapters {
		res, err := adapter.Detect(ctx, agentbridge.DetectEnv{})
		if err != nil {
			return fmt.Errorf("detect %s: %w", adapter.Name(), err)
		}
		detect := res
		entries = append(entries, providerEntryForAdapter(adapter, &detect))
	}
	return printJSON(struct {
		SchemaVersion string          `json:"schema_version"`
		Providers     []providerEntry `json:"providers"`
	}{
		SchemaVersion: BridgeDetectSchemaVersion,
		Providers:     entries,
	})
}

func providerEntryForAdapter(adapter agentbridge.Adapter, detect *agentbridge.DetectResult) providerEntry {
	return providerEntry{
		Name:              adapter.Name(),
		BlockedArgs:       adapter.BlockedArgs(),
		DefaultExecutable: providerDefaultExecutable(adapter.Name()),
		Detect:            detect,
	}
}
