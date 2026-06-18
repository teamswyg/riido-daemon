package codex

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
)

func codexCommandStartedEvent(p map[string]any) []agentbridge.Event {
	command := stringField(p, "command")
	return []agentbridge.Event{{
		Kind: agentbridge.EventToolCallStarted,
		Tool: agentbridge.ToolRef{
			ID:   stringField(p, "id"),
			Name: command,
			Kind: "shell",
			Args: toolargs.FromPairs("command", command),
		},
	}}
}

func codexCommandOutputEvent(p map[string]any) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventToolCallDelta,
		Tool: agentbridge.ToolRef{ID: stringField(p, "id")},
		Text: stringField(p, "chunk"),
	}}
}

func codexCommandCompletedEvent(p map[string]any) []agentbridge.Event {
	kind := agentbridge.EventToolCallCompleted
	if intField(p, "exit_code") != 0 {
		kind = agentbridge.EventToolCallFailed
	}
	return []agentbridge.Event{{
		Kind: kind,
		Tool: agentbridge.ToolRef{ID: stringField(p, "id"), Kind: "shell"},
	}}
}

func codexApplyPatchStartedEvent(p map[string]any) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventToolCallStarted,
		Tool: agentbridge.ToolRef{
			ID:   stringField(p, "id"),
			Kind: "patch_apply",
			Args: toolargs.FromPairs("path", stringField(p, "path")),
		},
	}}
}

func codexApplyPatchDoneEvent(p map[string]any) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventToolCallCompleted,
		Tool: agentbridge.ToolRef{ID: stringField(p, "id"), Kind: "patch_apply"},
	}}
}
