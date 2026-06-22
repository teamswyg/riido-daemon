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
//   - The installed CLI does not accept --model for agent runs. Model
//     selection is owned by OpenClaw config, for example `openclaw models set`.
package openclaw
