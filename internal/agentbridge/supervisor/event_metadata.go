package supervisor

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func firstMetadata(meta map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := meta[key]; value != "" {
			return value
		}
	}
	return ""
}

func runtimeIdentity(meta map[string]string) string {
	if value := meta[MetadataAgentIdentity]; value != "" {
		return value
	}
	if name := meta[MetadataAgentName]; name != "" {
		return "You are: " + name
	}
	return ""
}

func runtimeHardRules(meta map[string]string) []string {
	if meta == nil || strings.TrimSpace(meta[agentbridge.MetadataTelemetryContract]) == "" {
		return nil
	}
	return agentbridge.TelemetryNativeConfigHardRules()
}
