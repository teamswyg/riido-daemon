package codex

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
)

func translateServerRequest(method codexMethod, p map[string]any) []agentbridge.Event {
	switch method {
	case codexMethodApproveCommand:
		return codexApproveCommandEvent(p)
	case codexMethodApprovePatch:
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
			ID:   stringField(p, "id"),
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
			ID:   stringField(p, "id"),
			Name: path,
			Kind: "patch_apply",
			Args: toolargs.FromPairs("path", path),
		},
	}}
}
