// Package openclaw owns the C4 run-scope adapter for the OpenClaw CLI.
//
// Spawn shape:
//
//	openclaw agent exec --json <prompt>
//
// OpenClaw is the volatile one: flag sets can change between versions, so:
//   - agent exec owns isolated state so a running Gateway cannot collide.
//   - When the caller passes a SystemPrompt, we inline it into --message
//     because not every OpenClaw build supports --system-prompt.
//   - Model and cwd selection are applied through the task-scoped config and
//     process directory instead of caller-controlled protocol flags.
package openclaw
