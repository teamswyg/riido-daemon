package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/claude"
	"github.com/teamswyg/riido-daemon/internal/provider/codex"
	"github.com/teamswyg/riido-daemon/internal/provider/cursor"
	"github.com/teamswyg/riido-daemon/internal/provider/openclaw"
)

// BridgeProvidersSchemaVersion identifies the JSON shape printed by
// `riido bridge providers`.
const BridgeProvidersSchemaVersion = "riido-bridge-providers.v1"

// BridgeDetectSchemaVersion identifies the JSON shape printed by
// `riido bridge detect`.
const BridgeDetectSchemaVersion = "riido-bridge-detect.v1"

// providerEntry is the JSON-printable view of a registered adapter.
type providerEntry struct {
	Name              string                    `json:"name"`
	BlockedArgs       []string                  `json:"blocked_args"`
	DefaultExecutable string                    `json:"default_executable"`
	Detect            *agentbridge.DetectResult `json:"detect,omitempty"`
}

// registeredAdapters returns the canonical set of agentbridge adapters
// that ship with riido-daemon. Order is deterministic.
func registeredAdapters() []agentbridge.Adapter {
	return []agentbridge.Adapter{
		bridgeClaudeAdapter{},
		bridgeCodexAdapter{},
		bridgeOpenClawAdapter{},
		bridgeCursorAdapter{},
	}
}

// providerDefaultExecutable returns the binary name an adapter looks up
// on $PATH when no explicit override is given.
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

func runBridge(args []string) error {
	if len(args) < 1 {
		printUsage()
		return fmt.Errorf("missing bridge subcommand")
	}
	switch args[0] {
	case "providers":
		return runBridgeProviders(args[1:])
	case "detect":
		return runBridgeDetect(args[1:])
	default:
		printUsage()
		return fmt.Errorf("unknown bridge subcommand: %s", args[0])
	}
}

func runBridgeProviders(_ []string) error {
	entries := make([]providerEntry, 0, len(registeredAdapters()))
	for _, a := range registeredAdapters() {
		entries = append(entries, providerEntry{
			Name:              a.Name(),
			BlockedArgs:       a.BlockedArgs(),
			DefaultExecutable: providerDefaultExecutable(a.Name()),
		})
	}
	return printJSON(struct {
		SchemaVersion string          `json:"schema_version"`
		Providers     []providerEntry `json:"providers"`
	}{
		SchemaVersion: BridgeProvidersSchemaVersion,
		Providers:     entries,
	})
}

func runBridgeDetect(_ []string) error {
	ctx := context.Background()
	entries := make([]providerEntry, 0, len(registeredAdapters()))
	for _, a := range registeredAdapters() {
		res, err := a.Detect(ctx, agentbridge.DetectEnv{})
		if err != nil {
			return fmt.Errorf("detect %s: %w", a.Name(), err)
		}
		detect := res
		entries = append(entries, providerEntry{
			Name:              a.Name(),
			BlockedArgs:       a.BlockedArgs(),
			DefaultExecutable: providerDefaultExecutable(a.Name()),
			Detect:            &detect,
		})
	}
	return printJSON(struct {
		SchemaVersion string          `json:"schema_version"`
		Providers     []providerEntry `json:"providers"`
	}{
		SchemaVersion: BridgeDetectSchemaVersion,
		Providers:     entries,
	})
}

// --- Thin adapter wrappers so each provider exposes the Adapter
// interface uniformly. Full Detect/BuildStart/Parser wiring still
// belongs to each provider's package; these wrappers route the
// Adapter port through to the package-level helpers.

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
// `claude -p --input-format stream-json` blocks on stdin and the
// session hits the hard timeout.
func (bridgeClaudeAdapter) NewProtocolDriver(req agentbridge.StartRequest) (agentbridge.ProtocolDriver, error) {
	return claude.NewProtocolDriver(req)
}

type bridgeCodexAdapter struct{}

func (bridgeCodexAdapter) Name() string { return codex.Name }
func (bridgeCodexAdapter) Detect(ctx context.Context, env agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return codex.Detect(ctx, env)
}
func (bridgeCodexAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return codex.BuildStart(req, codex.StartOptions{AuthHomeDenyPath: codexAuthHomeDenyPath(req)})
}
func (bridgeCodexAdapter) NewParser() agentbridge.Parser { return codex.NewParser() }
func (bridgeCodexAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return codex.Translate(raw)
}
func (bridgeCodexAdapter) BlockedArgs() []string { return codex.BlockedArgs() }

// NewProtocolDriver lets RuntimeActor install the Codex JSON-RPC
// handshake driver. Implements agentbridge.ProtocolDriverProvider.
func (bridgeCodexAdapter) NewProtocolDriver(req agentbridge.StartRequest) (agentbridge.ProtocolDriver, error) {
	return codex.NewProtocolDriver(req)
}

func codexAuthHomeDenyPath(req agentbridge.StartRequest) string {
	if req.Env != nil {
		if value := strings.TrimSpace(req.Env["CODEX_HOME"]); value != "" {
			return value
		}
		if value := strings.TrimSpace(req.Env["HOME"]); value != "" {
			return filepath.Join(value, ".codex")
		}
	}
	if value := strings.TrimSpace(os.Getenv("CODEX_HOME")); value != "" {
		return value
	}
	if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
		return filepath.Join(home, ".codex")
	}
	return ""
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
