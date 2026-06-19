package main

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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
