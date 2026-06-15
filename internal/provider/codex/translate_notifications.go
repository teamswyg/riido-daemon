package codex

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
)

func translateNotification(method string, p map[string]any) []agentbridge.Event {
	switch method {
	case "thread_started", "thread_resumed":
		return []agentbridge.Event{{Kind: agentbridge.EventSessionIdentified, SessionID: stringField(p, "thread_id")}}

	case "thread/started", "thread/resumed":
		return []agentbridge.Event{{Kind: agentbridge.EventSessionIdentified, SessionID: threadIDFromParams(p)}}

	case "turn_started", "turn/started":
		return []agentbridge.Event{{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}}

	case "agent_message":
		return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: stringField(p, "text")}}

	case "item/agentMessage/delta":
		return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: stringField(p, "delta")}}

	case "reasoning":
		return []agentbridge.Event{{Kind: agentbridge.EventThinkingDelta, Text: stringField(p, "text")}}

	case "command_execution_started":
		return []agentbridge.Event{{
			Kind: agentbridge.EventToolCallStarted,
			Tool: agentbridge.ToolRef{
				ID:   stringField(p, "id"),
				Name: stringField(p, "command"),
				Kind: "shell",
				Args: toolargs.FromPairs("command", stringField(p, "command")),
			},
		}}

	case "command_execution_output":
		return []agentbridge.Event{{
			Kind: agentbridge.EventToolCallDelta,
			Tool: agentbridge.ToolRef{ID: stringField(p, "id")},
			Text: stringField(p, "chunk"),
		}}

	case "command_execution_completed":
		kind := agentbridge.EventToolCallCompleted
		if intField(p, "exit_code") != 0 {
			kind = agentbridge.EventToolCallFailed
		}
		return []agentbridge.Event{{
			Kind: kind,
			Tool: agentbridge.ToolRef{ID: stringField(p, "id"), Kind: "shell"},
		}}

	case "apply_patch_started":
		return []agentbridge.Event{{
			Kind: agentbridge.EventToolCallStarted,
			Tool: agentbridge.ToolRef{
				ID:   stringField(p, "id"),
				Kind: "patch_apply",
				Args: toolargs.FromPairs("path", stringField(p, "path")),
			},
		}}

	case "apply_patch_completed":
		return []agentbridge.Event{{
			Kind: agentbridge.EventToolCallCompleted,
			Tool: agentbridge.ToolRef{ID: stringField(p, "id"), Kind: "patch_apply"},
		}}

	case "turn_completed", "turn/completed":
		return []agentbridge.Event{{
			Kind: agentbridge.EventResult,
			Result: agentbridge.Result{
				Status: agentbridge.ResultCompleted,
				Output: stringField(p, "output"),
			},
		}}

	case "turn_error", "turn/error", "turn/failed":
		return []agentbridge.Event{{
			Kind: agentbridge.EventResult,
			Result: agentbridge.Result{
				Status: agentbridge.ResultFailed,
				Error:  stringField(p, "message"),
			},
		}}

	case "account/rateLimits/updated", "account_rate_limits_updated":
		// Codex app-server periodically reports current account rate-limit
		// windows. Informational (not terminal) — surface as a clear log so it
		// is no longer reported as an "unknown notification".
		return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "codex rate limits updated"}}

	// Newer codex app-server (0.13x) "item" + lifecycle notifications. The
	// assistant text still arrives via item/agentMessage/delta (handled above);
	// these are structural/lifecycle signals we acknowledge as informational so
	// they stop surfacing as "unknown notification" noise. Completion is NOT
	// inferred here — that stays with turn/completed and thread/status/changed
	// so a per-item completion can never truncate a live turn.
	case "item/started", "item/updated", "item/completed",
		"hook/started", "hook/completed",
		"mcpServer/startupStatus/updated",
		"remoteControl/status/changed":
		return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "codex " + method}}

	case "usage":
		return []agentbridge.Event{{Kind: agentbridge.EventUsageDelta, Usage: agentbridge.Usage{
			PromptTokens:     intField(p, "input_tokens"),
			CompletionTokens: intField(p, "output_tokens"),
			ReasoningTokens:  intField(p, "reasoning_tokens"),
		}}}

	case "thread/tokenUsage/updated":
		total := mapField(mapField(p, "tokenUsage"), "total")
		return []agentbridge.Event{{Kind: agentbridge.EventUsageDelta, Usage: agentbridge.Usage{
			PromptTokens:     intField(total, "inputTokens"),
			CompletionTokens: intField(total, "outputTokens"),
			ReasoningTokens:  intField(total, "reasoningOutputTokens"),
			CacheReadTokens:  intField(total, "cachedInputTokens"),
		}}}
	}

	return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "codex unknown notification: " + method}}
}

func translateServerRequest(method string, p map[string]any) []agentbridge.Event {
	switch method {
	case "approve_command":
		return []agentbridge.Event{{
			Kind: agentbridge.EventToolApprovalNeeded,
			Tool: agentbridge.ToolRef{
				ID:   stringField(p, "id"),
				Name: stringField(p, "command"),
				Kind: "shell",
				Args: toolargs.FromPairs("command", stringField(p, "command")),
			},
		}}
	case "approve_patch":
		return []agentbridge.Event{{
			Kind: agentbridge.EventToolApprovalNeeded,
			Tool: agentbridge.ToolRef{
				ID:   stringField(p, "id"),
				Name: stringField(p, "path"),
				Kind: "patch_apply",
				Args: toolargs.FromPairs("path", stringField(p, "path")),
			},
		}}
	}
	return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "codex unknown server_request: " + method}}
}

func params(raw agentbridge.RawEvent) map[string]any {
	p, _ := raw.Payload["params"].(map[string]any)
	return p
}

func threadIDFromParams(p map[string]any) string {
	if id := stringField(p, "threadId"); id != "" {
		return id
	}
	if id := stringField(p, "thread_id"); id != "" {
		return id
	}
	thread := mapField(p, "thread")
	if id := stringField(thread, "id"); id != "" {
		return id
	}
	return stringField(thread, "sessionId")
}
