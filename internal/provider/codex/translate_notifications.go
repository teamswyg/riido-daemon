package codex

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
)

func translateNotification(method codexMethod, p map[string]any) []agentbridge.Event {
	switch method {
	case codexMethodThreadStarted, codexMethodThreadResumed:
		return []agentbridge.Event{{Kind: agentbridge.EventSessionIdentified, SessionID: stringField(p, "thread_id")}}

	case codexMethodThreadStartedSlash, codexMethodThreadResumedSlash:
		return []agentbridge.Event{{Kind: agentbridge.EventSessionIdentified, SessionID: threadIDFromParams(p)}}

	case codexMethodTurnStarted, codexMethodTurnStartedSlash:
		return []agentbridge.Event{{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}}

	case codexMethodAgentMessage:
		return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: stringField(p, "text")}}

	case codexMethodItemAgentMessageDelta:
		return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: stringField(p, "delta")}}

	case codexMethodReasoning:
		return []agentbridge.Event{{Kind: agentbridge.EventThinkingDelta, Text: stringField(p, "text")}}

	case codexMethodCommandStarted:
		return []agentbridge.Event{{
			Kind: agentbridge.EventToolCallStarted,
			Tool: agentbridge.ToolRef{
				ID:   stringField(p, "id"),
				Name: stringField(p, "command"),
				Kind: "shell",
				Args: toolargs.FromPairs("command", stringField(p, "command")),
			},
		}}

	case codexMethodCommandOutput:
		return []agentbridge.Event{{
			Kind: agentbridge.EventToolCallDelta,
			Tool: agentbridge.ToolRef{ID: stringField(p, "id")},
			Text: stringField(p, "chunk"),
		}}

	case codexMethodCommandCompleted:
		kind := agentbridge.EventToolCallCompleted
		if intField(p, "exit_code") != 0 {
			kind = agentbridge.EventToolCallFailed
		}
		return []agentbridge.Event{{
			Kind: kind,
			Tool: agentbridge.ToolRef{ID: stringField(p, "id"), Kind: "shell"},
		}}

	case codexMethodApplyPatchStart:
		return []agentbridge.Event{{
			Kind: agentbridge.EventToolCallStarted,
			Tool: agentbridge.ToolRef{
				ID:   stringField(p, "id"),
				Kind: "patch_apply",
				Args: toolargs.FromPairs("path", stringField(p, "path")),
			},
		}}

	case codexMethodApplyPatchDone:
		return []agentbridge.Event{{
			Kind: agentbridge.EventToolCallCompleted,
			Tool: agentbridge.ToolRef{ID: stringField(p, "id"), Kind: "patch_apply"},
		}}

	case codexMethodTurnCompleted, codexMethodTurnCompleteSlash:
		return []agentbridge.Event{{
			Kind: agentbridge.EventResult,
			Result: agentbridge.Result{
				Status: agentbridge.ResultCompleted,
				Output: stringField(p, "output"),
			},
		}}

	case codexMethodTurnError, codexMethodTurnErrorSlash, codexMethodTurnFailedSlash:
		return []agentbridge.Event{{
			Kind: agentbridge.EventResult,
			Result: agentbridge.Result{
				Status: agentbridge.ResultFailed,
				Error:  stringField(p, "message"),
			},
		}}

	case codexMethodAccountRateLimitsUpdated, codexMethodAccountRateLimitsAlt:
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
	case codexMethodItemStarted, codexMethodItemUpdated, codexMethodItemCompleted,
		codexMethodHookStarted, codexMethodHookCompleted,
		codexMethodMCPStartupStatusUpdated,
		codexMethodRemoteControlChanged:
		return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "codex " + string(method)}}

	case codexMethodUsage:
		return []agentbridge.Event{{Kind: agentbridge.EventUsageDelta, Usage: agentbridge.Usage{
			PromptTokens:     intField(p, "input_tokens"),
			CompletionTokens: intField(p, "output_tokens"),
			ReasoningTokens:  intField(p, "reasoning_tokens"),
		}}}

	case codexMethodThreadTokenUsage:
		total := mapField(mapField(p, "tokenUsage"), "total")
		return []agentbridge.Event{{Kind: agentbridge.EventUsageDelta, Usage: agentbridge.Usage{
			PromptTokens:     intField(total, "inputTokens"),
			CompletionTokens: intField(total, "outputTokens"),
			ReasoningTokens:  intField(total, "reasoningOutputTokens"),
			CacheReadTokens:  intField(total, "cachedInputTokens"),
		}}}
	default:
		return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "codex unknown notification: " + string(method)}}
	}
}

func translateServerRequest(method codexMethod, p map[string]any) []agentbridge.Event {
	switch method {
	case codexMethodApproveCommand:
		return []agentbridge.Event{{
			Kind: agentbridge.EventToolApprovalNeeded,
			Tool: agentbridge.ToolRef{
				ID:   stringField(p, "id"),
				Name: stringField(p, "command"),
				Kind: "shell",
				Args: toolargs.FromPairs("command", stringField(p, "command")),
			},
		}}
	case codexMethodApprovePatch:
		return []agentbridge.Event{{
			Kind: agentbridge.EventToolApprovalNeeded,
			Tool: agentbridge.ToolRef{
				ID:   stringField(p, "id"),
				Name: stringField(p, "path"),
				Kind: "patch_apply",
				Args: toolargs.FromPairs("path", stringField(p, "path")),
			},
		}}
	default:
		return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "codex unknown server_request: " + string(method)}}
	}
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
