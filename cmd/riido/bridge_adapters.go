package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/claude"
	"github.com/teamswyg/riido-daemon/internal/provider/codex"
	"github.com/teamswyg/riido-daemon/internal/provider/cursor"
	"github.com/teamswyg/riido-daemon/internal/provider/openclaw"
)

// builtinAgentAdapters returns the canonical provider adapter set used by both
// the bridge CLI and the daemon runtime. Order is deterministic.
func builtinAgentAdapters() []agentbridge.Adapter {
	return []agentbridge.Adapter{
		bridgeClaudeAdapter{},
		bridgeCodexAdapter{},
		bridgeOpenClawAdapter{},
		bridgeCursorAdapter{},
	}
}

// providerDefaultExecutable returns the binary name an adapter looks up on
// $PATH when no explicit override is given.
func providerDefaultExecutable(name string) string {
	switch name {
	case claude.Name:
		return claude.DefaultExecutable
	case codex.Name:
		return codex.DefaultExecutable
	case openclaw.Name:
		return openclaw.DefaultExecutable
	case cursor.Name:
		return cursor.DefaultExecutable
	}
	return ""
}

// Thin adapter wrappers expose each provider package through the common
// agentbridge.Adapter port. Provider-specific behavior stays in the provider
// packages; this file is only the built-in wiring table.
type bridgeClaudeAdapter struct{}

func (bridgeClaudeAdapter) Name() string { return claude.Name }
func (bridgeClaudeAdapter) Detect(ctx context.Context, env agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return claude.Detect(ctx, env)
}

func (bridgeClaudeAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return claude.BuildStart(req, claude.StartOptions{PermissionMode: claude.PermissionModeApproval})
}
func (bridgeClaudeAdapter) NewParser() agentbridge.Parser { return claude.NewParser() }
func (bridgeClaudeAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return claude.Translate(raw)
}
func (bridgeClaudeAdapter) BlockedArgs() []string { return claude.BlockedArgs() }
func (bridgeClaudeAdapter) BuildProviderInput(cmd agentbridge.Command) ([]byte, error) {
	return claude.BuildProviderInput(cmd)
}

// NewProtocolDriver wires the Claude stdin-frame driver. Without it,
// `claude -p --input-format stream-json` blocks on stdin and the session hits
// the hard timeout.
func (bridgeClaudeAdapter) NewProtocolDriver(req agentbridge.StartRequest) (agentbridge.ProtocolDriver, error) {
	return claude.NewProtocolDriver(req)
}

type bridgeCodexAdapter struct{}

func (bridgeCodexAdapter) Name() string { return codex.Name }
func (bridgeCodexAdapter) Detect(ctx context.Context, env agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return codex.Detect(ctx, env)
}

func (bridgeCodexAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return codex.BuildStart(req, codex.StartOptions{})
}
func (bridgeCodexAdapter) NewParser() agentbridge.Parser { return codex.NewParser() }
func (bridgeCodexAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return codex.Translate(raw)
}
func (bridgeCodexAdapter) BlockedArgs() []string { return codex.BlockedArgs() }

// NewProtocolDriver lets RuntimeActor install the Codex JSON-RPC handshake
// driver. Implements agentbridge.ProtocolDriverProvider.
func (bridgeCodexAdapter) NewProtocolDriver(req agentbridge.StartRequest) (agentbridge.ProtocolDriver, error) {
	return codex.NewProtocolDriver(req)
}

type bridgeOpenClawAdapter struct{}

func (bridgeOpenClawAdapter) Name() string { return openclaw.Name }
func (bridgeOpenClawAdapter) Detect(ctx context.Context, env agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return openclaw.Detect(ctx, env)
}

func (bridgeOpenClawAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return openclaw.BuildStart(req, openclaw.StartOptions{})
}
func (bridgeOpenClawAdapter) NewParser() agentbridge.Parser { return openclaw.NewParser() }
func (bridgeOpenClawAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return openclaw.Translate(raw)
}
func (bridgeOpenClawAdapter) BlockedArgs() []string { return openclaw.BlockedArgs() }

type bridgeCursorAdapter struct{}

func (bridgeCursorAdapter) Name() string { return cursor.Name }
func (bridgeCursorAdapter) Detect(ctx context.Context, env agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return cursor.Detect(ctx, env)
}

func (bridgeCursorAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return cursor.BuildStart(req, cursor.StartOptions{})
}
func (bridgeCursorAdapter) NewParser() agentbridge.Parser { return cursor.NewParser() }
func (bridgeCursorAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return cursor.Translate(raw)
}
func (bridgeCursorAdapter) BlockedArgs() []string { return cursor.BlockedArgs() }
