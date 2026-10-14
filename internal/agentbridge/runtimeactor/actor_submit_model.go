package runtimeactor

import (
	"strings"

	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func submitModel(provider bridge.Provider, requested, resumeSessionID string, models []RuntimeModel) string {
	requested = strings.TrimSpace(requested)
	if requested != "" || strings.TrimSpace(resumeSessionID) != "" || !providercatalog.IsCodex(string(provider)) {
		return requested
	}
	for _, model := range models {
		if model.IsDefault {
			return strings.TrimSpace(model.ModelID)
		}
	}
	return ""
}
