// Package codex owns the C4 run-scope adapter for OpenAI's Codex CLI. The
// spawn shape is `codex --sandbox danger-full-access app-server --listen
// stdio://` so the session actor speaks JSON-RPC 2.0 over the process's stdio
// pipes while the daemon, not provider defaults or caller args, owns execution
// authority.
//
// What this slice provides:
//   - Command builder with protocol-critical args locked in.
//   - --listen blocklist (caller cannot reroute transport).
//   - daemon-owned full-access sandbox selection.
//   - JSON-RPC protocol driver via the provider-neutral agentbridge port.
package codex

import (
	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const (
	Name                  = string(providercatalog.KindCodex)
	DefaultExecutable     = "codex"
	FullAccessSandboxMode = "danger-full-access"
)

// BlockedArgs are protocol-critical flags the adapter sets itself.
// --listen is the load-bearing one: caller-supplied --listen would let
// callers reroute Codex onto an arbitrary transport, breaking the
// adapter's JSON-RPC-over-stdio contract.
func BlockedArgs() []string {
	return providercap.ProtocolCriticalArgs(providercap.ProtocolCodexAppServer)
}

// BuildStart turns an agentbridge.StartRequest + Codex options into a
// runtime.StartCommand.
func BuildStart(req agentbridge.StartRequest, opts StartOptions) (agentbridge.StartCommand, error) {
	exe := opts.Executable
	if exe == "" {
		exe = req.Executable
	}
	if exe == "" {
		exe = DefaultExecutable
	}

	args := []string{
		"--sandbox", FullAccessSandboxMode,
		"app-server",
	}
	args = append(args, "--listen", "stdio://")

	kept, dropped := filterCustomArgs(req.CustomArgs)
	args = append(args, kept...)

	env := buildEnv(req.Env, opts)

	return agentbridge.StartCommand{
		Executable:  exe,
		Args:        args,
		Env:         env,
		Dir:         req.Cwd,
		StdinMode:   agentbridge.StdinPipe,
		DroppedArgs: dropped,
	}, nil
}
