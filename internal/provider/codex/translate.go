package codex

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Translate maps a Codex JSON-RPC RawEvent to run-scope Events.
//
// Notification methods we recognize (params are conventionally a map):
//   - thread_started / thread_resumed → SessionIdentified
//   - turn_started → Lifecycle(Running)
//   - agent_message → TextDelta
//   - reasoning → ThinkingDelta
//   - command_execution_started → ToolCallStarted(kind=shell)
//   - command_execution_output → ToolCallDelta
//   - command_execution_completed → ToolCallCompleted / Failed (by exit_code)
//   - apply_patch_started / apply_patch_completed → ToolCallStarted/Completed(kind=patch_apply)
//   - turn_completed → Result(completed)
//   - turn_error → Result(failed)
//   - usage → UsageDelta
//
// Server-initiated requests (method present, id present):
//   - approve_command / approve_patch → ToolApprovalNeeded
//
// Anything else surfaces as Log so we don't silently drop it
// (spec §15 item 3).
func Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	switch raw.Source {
	case agentbridge.RawSourceStderr:
		return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: string(raw.Bytes)}}, nil, nil
	case agentbridge.RawSourceStdout, agentbridge.RawSourceClose:
	}

	switch {
	case raw.Type == "malformed":
		return []agentbridge.Event{{Kind: agentbridge.EventWarning, Text: "malformed codex json-rpc frame", Err: string(raw.Bytes)}}, nil, nil

	case raw.Type == "error":
		return []agentbridge.Event{{
			Kind: agentbridge.EventError,
			Err:  errMessage(raw.Payload),
		}}, nil, nil

	case raw.Type == "response":
		// Plain RPC responses aren't run-scope events; the RPC actor
		// resolves them. Emit a Log so observability is preserved.
		return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "codex rpc response"}}, nil, nil

	case strings.HasPrefix(raw.Type, "notification:"):
		method := strings.TrimPrefix(raw.Type, "notification:")
		return translateNotification(method, params(raw)), nil, nil

	case strings.HasPrefix(raw.Type, "server_request:"):
		method := strings.TrimPrefix(raw.Type, "server_request:")
		return translateServerRequest(method, params(raw)), nil, nil
	}

	return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "codex unknown frame: " + raw.Type}}, nil, nil
}
