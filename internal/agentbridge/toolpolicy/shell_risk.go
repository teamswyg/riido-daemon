package toolpolicy

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

func classifyShellRisk(args map[string]string) (policy.ToolUseSurface, bool) {
	command, ok := commandArg(args)
	if !ok || strings.TrimSpace(command) == "" {
		return policy.ToolUseDestructiveCommand, true
	}
	switch {
	case commandContainsNetworkEgress(command):
		return policy.ToolUseNetworkEgress, true
	case commandExposesSecrets(command):
		return policy.ToolUseSecretExposure, true
	case commandIsDestructive(command):
		return policy.ToolUseDestructiveCommand, true
	case commandTouchesProtectedPath(command):
		return policy.ToolUseProtectedPathWrite, true
	default:
		return "", false
	}
}
