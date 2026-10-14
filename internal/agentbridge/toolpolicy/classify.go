package toolpolicy

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func ClassifyToolUseSurface(tool agentbridge.ToolRef) (policy.ToolUseSurface, bool) {
	kind := normalizeToolToken(tool.Kind)
	name := normalizeToolToken(tool.Name)
	switch {
	case classifiesSecretExposure(kind, name, tool.Args):
		return policy.ToolUseSecretExposure, true
	case classifiesNetworkEgress(kind, name, tool.Args):
		return policy.ToolUseNetworkEgress, true
	case classifiesProtectedPathWrite(kind, name, tool.Args):
		return policy.ToolUseProtectedPathWrite, true
	case classifiesShellRisk(kind, name, tool.Args):
		return classifyShellRisk(tool.Args)
	default:
		return "", false
	}
}

func classifiesSecretExposure(kind, name string, args map[string]string) bool {
	return matchesAny(kind, name, "secret", "secrets", "token", "tokens", "credential", "credentials") ||
		hasSensitiveArgSignal(args)
}

func classifiesNetworkEgress(kind, name string, args map[string]string) bool {
	return matchesAny(kind, name, "webfetch", "web_fetch", "websearch", "web_search", "fetch", "http", "network") ||
		argsContainNetworkEgress(args)
}

func classifiesProtectedPathWrite(kind, name string, args map[string]string) bool {
	if !matchesAny(kind, name, "patch_apply", "apply_patch", "edit", "write", "multiedit", "multi_edit", "delete", "remove") {
		return false
	}
	return len(args) == 0 || argsTouchProtectedPath(args)
}

func classifiesShellRisk(kind, name string, args map[string]string) bool {
	if !matchesAny(kind, name, "shell", "bash", "exec", "command", "run_command", "terminal") {
		return false
	}
	command, ok := commandArg(args)
	return !ok || strings.TrimSpace(command) == "" ||
		commandContainsNetworkEgress(command) ||
		commandExposesSecrets(command) ||
		commandIsDestructive(command) ||
		commandTouchesProtectedPath(command)
}
