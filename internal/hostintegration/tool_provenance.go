package hostintegration

// ToolProvenance records why Riido trusts an executable path enough to show it
// as a provider candidate.
type ToolProvenance string

const (
	ToolProvenanceUserSelected ToolProvenance = "user-selected"
	ToolProvenanceEnvOverride  ToolProvenance = "env-override"
	ToolProvenanceAutoDetected ToolProvenance = "auto-detected"
)

// Valid reports whether provenance is one of the SSOT-defined provenance values.
func (p ToolProvenance) Valid() bool {
	switch p {
	case ToolProvenanceUserSelected, ToolProvenanceEnvOverride, ToolProvenanceAutoDetected:
		return true
	default:
		return false
	}
}

func provenanceRank(provenance ToolProvenance) int {
	switch provenance {
	case ToolProvenanceUserSelected:
		return 3
	case ToolProvenanceEnvOverride:
		return 2
	case ToolProvenanceAutoDetected:
		return 1
	default:
		return 0
	}
}
