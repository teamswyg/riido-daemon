// Package openclaw owns the C4 run-scope adapter for the OpenClaw CLI.
//
// Spawn shape:
//
//	openclaw agent --local --json --session-id <id> --message <prompt>
//
// OpenClaw is the volatile one: flag sets can change between versions, so:
//   - We require an explicit session id. StartOptions.SessionID wins;
//     otherwise ResolveSessionID maps provider-neutral ResumeSessionID
//     or TaskID to --session-id. Empty fallback is never allowed.
//   - When the caller passes a SystemPrompt, we inline it into --message
//     because not every OpenClaw build supports --system-prompt.
//   - Model is treated as an agent/profile name, not an LLM identifier.
package openclaw

import (
	"errors"
	"fmt"
	"strings"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const Name = "openclaw"
const DefaultExecutable = "openclaw"

// BlockedArgs lists the protocol-critical flags this adapter sets itself.
// Callers cannot override these via CustomArgs.
func BlockedArgs() []string {
	return providercap.ProtocolCriticalArgs(providercap.ProtocolOpenClawAgentJSON)
}

type StartOptions struct {
	// Executable overrides the binary path.
	Executable string
	// SessionID overrides the provider-neutral session id resolution.
	// OpenClaw's resume model is session-id-based; silently using an
	// empty session id would create an anonymous run.
	SessionID string
}

func BuildStart(req agentbridge.StartRequest, opts StartOptions) (agentbridge.StartCommand, error) {
	sessionID := opts.SessionID
	if sessionID == "" {
		var err error
		sessionID, err = ResolveSessionID(req)
		if err != nil {
			return agentbridge.StartCommand{}, err
		}
	}
	exe := opts.Executable
	if exe == "" {
		exe = DefaultExecutable
	}

	message := buildMessage(req.SystemPrompt, req.Prompt)

	args := []string{
		"agent",
		"--local",
		"--json",
		"--session-id", sessionID,
	}
	if req.Model != "" {
		args = append(args, "--model", req.Model)
	}
	args = append(args, "--message", message)

	kept, dropped := agentbridge.FilterBlockedArgs(req.CustomArgs, BlockedArgs())
	args = append(args, kept...)

	env := make([]string, 0, len(req.Env))
	for k, v := range req.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return agentbridge.StartCommand{
		Executable:  exe,
		Args:        args,
		Env:         env,
		Dir:         req.Cwd,
		StdinMode:   agentbridge.StdinNone,
		DroppedArgs: dropped,
	}, nil
}

// ResolveSessionID maps the provider-neutral start request onto
// OpenClaw's mandatory --session-id. ResumeSessionID preserves a
// provider session, while TaskID gives first-run tasks a deterministic
// session id without inventing provider state.
func ResolveSessionID(req agentbridge.StartRequest) (string, error) {
	if strings.TrimSpace(req.ResumeSessionID) != "" {
		return req.ResumeSessionID, nil
	}
	if strings.TrimSpace(req.TaskID) != "" {
		return req.TaskID, nil
	}
	return "", errors.New("openclaw: SessionID is required (set ResumeSessionID or TaskID)")
}

// buildMessage inlines the system prompt above the user prompt when both
// are present, separated by a blank line. When system prompt is empty,
// the user prompt is returned verbatim.
func buildMessage(systemPrompt, userPrompt string) string {
	system := strings.TrimSpace(systemPrompt)
	user := strings.TrimSpace(userPrompt)
	if system == "" {
		return userPrompt
	}
	return system + "\n\n" + user
}
