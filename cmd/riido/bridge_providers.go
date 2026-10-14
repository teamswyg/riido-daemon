package main

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
