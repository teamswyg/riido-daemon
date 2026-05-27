// Package codex owns the C4 run-scope adapter for OpenAI's Codex CLI. The
// spawn shape is `codex app-server --listen stdio://` so the session actor
// speaks JSON-RPC 2.0 over the process's stdio pipes.
//
// What this slice provides:
//   - Command builder with protocol-critical args locked in.
//   - --listen blocklist (caller cannot reroute transport).
//   - CODEX_HOME per-task isolation (caller's env cannot override).
//   - JSON-RPC protocol driver via the provider-neutral agentbridge port.
package codex

import (
	"fmt"
	"maps"
	"sort"
	"strings"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const Name = "codex"
const DefaultExecutable = "codex"

// BlockedArgs are protocol-critical flags the adapter sets itself.
// --listen is the load-bearing one: caller-supplied --listen would let
// callers reroute Codex onto an arbitrary transport, breaking the
// adapter's JSON-RPC-over-stdio contract.
func BlockedArgs() []string {
	return providercap.ProtocolCriticalArgs(providercap.ProtocolCodexAppServer)
}

// UnsafeBypassArgs are Codex flags covered by docs/20-domain/security.md §5.
// BuildStart does not expose a policy-approved allow path for these surfaces,
// so free-form CustomArgs must not smuggle them into the provider process.
// Boolean equals-forms such as --yolo=true are the same unsafe surface.
func UnsafeBypassArgs() []string {
	return []string{
		"--yolo",
		"--dangerously-bypass-approvals-and-sandbox",
		"--sandbox=danger-full-access",
	}
}

// StartOptions carries Codex-specific knobs.
type StartOptions struct {
	// Executable overrides the binary path. Falls back to DefaultExecutable.
	Executable string
	// CodexHome is the per-task isolated $CODEX_HOME directory. When
	// non-empty it is injected into the process env and the caller
	// cannot override it. See docs/20-domain/provider-runtime.md.
	CodexHome string
}

// BuildStart turns an agentbridge.StartRequest + Codex options into a
// runtime.StartCommand.
func BuildStart(req agentbridge.StartRequest, opts StartOptions) (agentbridge.StartCommand, error) {
	exe := opts.Executable
	if exe == "" {
		exe = DefaultExecutable
	}

	args := []string{
		"app-server",
		"--listen", "stdio://",
	}

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

func filterCustomArgs(custom []string) (kept []string, dropped []string) {
	kept, dropped = agentbridge.FilterBlockedArgs(custom, BlockedArgs())
	return filterUnsafeBypassArgs(kept, dropped)
}

func filterUnsafeBypassArgs(custom []string, dropped []string) (kept []string, allDropped []string) {
	blocked := make(map[string]struct{}, len(UnsafeBypassArgs()))
	for _, arg := range UnsafeBypassArgs() {
		blocked[arg] = struct{}{}
	}

	allDropped = append(allDropped, dropped...)
	for i := 0; i < len(custom); i++ {
		arg := custom[i]
		if _, isBlocked := blocked[arg]; isBlocked {
			allDropped = append(allDropped, arg)
			continue
		}
		if strings.HasPrefix(arg, "--yolo=") ||
			strings.HasPrefix(arg, "--dangerously-bypass-approvals-and-sandbox=") {
			allDropped = append(allDropped, arg)
			continue
		}
		if strings.HasPrefix(arg, "--sandbox=") {
			if strings.TrimPrefix(arg, "--sandbox=") == "danger-full-access" {
				allDropped = append(allDropped, arg)
				continue
			}
		}
		if arg == "--sandbox" && i+1 < len(custom) && custom[i+1] == "danger-full-access" {
			allDropped = append(allDropped, arg, custom[i+1])
			i++
			continue
		}
		kept = append(kept, arg)
	}
	return kept, allDropped
}

// buildEnv merges caller env with adapter-reserved env. Reserved keys
// always win — caller values for those keys are silently dropped. We
// don't surface the drop as a Warning event because env collisions are
// expected (the daemon may pass through user $PATH etc. that contains
// no secrets, and adding a warning per env collision would be noisy).
// If we later need observability, switch to returning a separate
// DroppedEnvKeys slice.
func buildEnv(caller map[string]string, opts StartOptions) []string {
	reserved := map[string]string{}
	if opts.CodexHome != "" {
		reserved["CODEX_HOME"] = opts.CodexHome
	}

	merged := make(map[string]string, len(caller)+len(reserved))
	for k, v := range caller {
		if _, isReserved := reserved[k]; isReserved {
			continue
		}
		merged[k] = v
	}
	maps.Copy(merged, reserved)

	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	env := make([]string, 0, len(keys))
	for _, k := range keys {
		env = append(env, fmt.Sprintf("%s=%s", k, merged[k]))
	}
	return env
}
