// Package claude owns the C4 run-scope adapter for Anthropic's Claude Code CLI.
// It owns command construction, executable detection, stream-json parsing,
// translation, and stdin protocol framing. The adapter is a translator; it does
// NOT own a state machine of its own. agentbridge does.
//
// This package provides:
//   - The blocked-args list (protocol-critical flags the adapter sets itself).
//   - BuildStart: turns an agentbridge.StartRequest into a StartCommand.
//   - An explicit, required PermissionMode parameter. There is no default that
//     maps to bypassPermissions. See docs/20-domain/security.md.
//   - Detect/NewParser/Translate/NewProtocolDriver adapter hooks.
package claude
