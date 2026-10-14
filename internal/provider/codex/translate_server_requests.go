package codex

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
)

func translateServerRequest(method codexMethod, p map[string]any) []agentbridge.Event {
	switch method {
	case codexMethod("item/tool/requestUserInput"):
		return []agentbridge.Event{{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateWaitingUserInput, Tool: agentbridge.ToolRef{ID: firstApprovalID(p), Kind: "user_input"}}}
	case codexMethodApproveCommand, codexMethod("item/commandExecution/requestApproval"):
		return codexApproveCommandEvent(p)
	case codexMethodApprovePatch, codexMethod("item/fileChange/requestApproval"):
		return codexApprovePatchEvent(p)
	default:
		return []agentbridge.Event{{
			Kind: agentbridge.EventLog,
			Text: "codex unknown server_request: " + string(method),
		}}
	}
}

func codexApproveCommandEvent(p map[string]any) []agentbridge.Event {
	command := stringField(p, "command")
	return []agentbridge.Event{{
		Kind: agentbridge.EventToolApprovalNeeded,
		Tool: agentbridge.ToolRef{
			ID:   firstApprovalID(p),
			Name: command,
			Kind: "shell",
			Args: toolargs.FromPairs("command", command),
		},
	}}
}

func codexApprovePatchEvent(p map[string]any) []agentbridge.Event {
	path := stringField(p, "path")
	return []agentbridge.Event{{
		Kind: agentbridge.EventToolApprovalNeeded,
		Tool: agentbridge.ToolRef{
			ID:   firstApprovalID(p),
			Name: path,
			Kind: "patch_apply",
			Args: toolargs.FromPairs("path", path),
		},
	}}
}

func firstApprovalID(p map[string]any) string {
	if id := stringField(p, "itemId"); id != "" {
		return id
	}
	return stringField(p, "id")
}
