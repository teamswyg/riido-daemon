package codex

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
)

func codexCommandExecutionApprovalEvent(p map[string]any, reqID string) []agentbridge.Event {
	command := stringField(p, "command")
	return []agentbridge.Event{{
		Kind: agentbridge.EventToolApprovalNeeded,
		Tool: agentbridge.ToolRef{
			ID:                firstNonEmpty(stringField(p, "approvalId"), stringField(p, "itemId"), reqID),
			Name:              firstNonEmpty(command, stringField(p, "itemId")),
			Kind:              "shell",
			Args:              toolargs.FromValue(p),
			ProviderRequestID: reqID,
		},
	}}
}

func codexFileChangeApprovalEvent(p map[string]any, reqID string) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventToolApprovalNeeded,
		Tool: agentbridge.ToolRef{
			ID:                firstNonEmpty(stringField(p, "itemId"), reqID),
			Name:              firstNonEmpty(stringField(p, "grantRoot"), stringField(p, "itemId"), "file change"),
			Kind:              "patch_apply",
			Args:              toolargs.FromValue(p),
			ProviderRequestID: reqID,
		},
	}}
}
