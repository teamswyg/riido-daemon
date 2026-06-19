package main

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

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

func providerEntryForAdapter(adapter agentbridge.Adapter, detect *agentbridge.DetectResult) providerEntry {
	return providerEntry{
		Name:              adapter.Name(),
		BlockedArgs:       adapter.BlockedArgs(),
		DefaultExecutable: providerDefaultExecutable(adapter.Name()),
		Detect:            detect,
	}
}
