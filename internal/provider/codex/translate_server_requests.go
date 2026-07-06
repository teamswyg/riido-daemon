package codex

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
)

func translateServerRequest(method codexMethod, payload map[string]any) []agentbridge.Event {
	p := paramsPayload(payload)
	reqID := providerRequestID(payload)
	switch method {
	case codexMethodApproveCommand:
		return codexApproveCommandEvent(p, reqID)
	case codexMethodApprovePatch:
		return codexApprovePatchEvent(p, reqID)
	case codexMethodCommandExecutionRequestApproval:
		return codexCommandExecutionApprovalEvent(p, reqID)
	case codexMethodFileChangeRequestApproval:
		return codexFileChangeApprovalEvent(p, reqID)
	default:
		return []agentbridge.Event{{
			Kind: agentbridge.EventLog,
			Text: "codex unknown server_request: " + string(method),
		}}
	}
}

func codexApproveCommandEvent(p map[string]any, reqID string) []agentbridge.Event {
	command := stringField(p, "command")
	return []agentbridge.Event{{
		Kind: agentbridge.EventToolApprovalNeeded,
		Tool: agentbridge.ToolRef{
			ID:                firstNonEmpty(stringField(p, "id"), reqID),
			Name:              command,
			Kind:              "shell",
			Args:              toolargs.FromPairs("command", command),
			ProviderRequestID: reqID,
		},
	}}
}

func codexApprovePatchEvent(p map[string]any, reqID string) []agentbridge.Event {
	path := stringField(p, "path")
	return []agentbridge.Event{{
		Kind: agentbridge.EventToolApprovalNeeded,
		Tool: agentbridge.ToolRef{
			ID:                firstNonEmpty(stringField(p, "id"), reqID),
			Name:              path,
			Kind:              "patch_apply",
			Args:              toolargs.FromPairs("path", path),
			ProviderRequestID: reqID,
		},
	}}
}
