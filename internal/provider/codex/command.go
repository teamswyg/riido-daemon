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
	"fmt"
	"maps"
	"sort"
	"strings"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const (
	Name                  = "codex"
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

// UnsafeBypassArgs are provider-native approval-bypass flags covered by
// docs/20-domain/security.md §5. The daemon does not expose an allow path for
// these free-form CustomArgs. Boolean equals-forms such as --yolo=true are the
// same unsafe surface.
//
// Codex `--sandbox danger-full-access` is deliberately not in this list: it is
// the daemon-owned provider full-access runtime envelope, not a caller-owned
// bypass flag.
func UnsafeBypassArgs() []string {
	return []string{
		"--yolo",
		"--dangerously-bypass-approvals-and-sandbox",
	}
}

// SandboxOverrideArgs are Codex sandbox-selection flags. The daemon owns the
// provider trust envelope, so caller CustomArgs may not override it.
func SandboxOverrideArgs() []string {
	return []string{"--sandbox", "-s"}
}

// SecurityCriticalArgs are Codex app-server flags that can rewrite the
// daemon-owned launch/trust shape. They are distinct from protocol-critical
// args: --listen protects transport shape, while these protect C4/C7 runtime
// policy decisions from caller-provided config overlays.
func SecurityCriticalArgs() []string {
	return []string{
		"-c",
		"--config",
		"--enable",
		"--disable",
	}
}

// StartOptions carries Codex-specific knobs.
type StartOptions struct {
	// Executable overrides the binary path. Falls back to DefaultExecutable.
	Executable string
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

func filterCustomArgs(custom []string) (kept, dropped []string) {
	kept, dropped = agentbridge.FilterBlockedArgs(custom, BlockedArgs())
	kept, dropped = filterConfigOverrideArgs(kept, dropped)
	kept, dropped = filterSandboxOverrideArgs(kept, dropped)
	return filterUnsafeBypassArgs(kept, dropped)
}

func filterConfigOverrideArgs(custom, dropped []string) (kept, allDropped []string) {
	blocked := make(map[string]struct{}, len(SecurityCriticalArgs()))
	for _, arg := range SecurityCriticalArgs() {
		blocked[arg] = struct{}{}
	}

	allDropped = append(allDropped, dropped...)
	for i := 0; i < len(custom); i++ {
		arg := custom[i]
		if _, isBlocked := blocked[arg]; isBlocked {
			allDropped = append(allDropped, arg)
			if (arg == "-c" || arg == "--config" || arg == "--enable" || arg == "--disable") && i+1 < len(custom) {
				allDropped = append(allDropped, custom[i+1])
				i++
			}
			continue
		}
		if strings.HasPrefix(arg, "-c=") ||
			strings.HasPrefix(arg, "--config=") ||
			strings.HasPrefix(arg, "--enable=") ||
			strings.HasPrefix(arg, "--disable=") {
			allDropped = append(allDropped, arg)
			continue
		}
		kept = append(kept, arg)
	}
	return kept, allDropped
}

func filterSandboxOverrideArgs(custom, dropped []string) (kept, allDropped []string) {
	blocked := make(map[string]struct{}, len(SandboxOverrideArgs()))
	for _, arg := range SandboxOverrideArgs() {
		blocked[arg] = struct{}{}
	}

	allDropped = append(allDropped, dropped...)
	for i := 0; i < len(custom); i++ {
		arg := custom[i]
		if _, isBlocked := blocked[arg]; isBlocked {
			allDropped = append(allDropped, arg)
			if i+1 < len(custom) {
				allDropped = append(allDropped, custom[i+1])
				i++
			}
			continue
		}
		if strings.HasPrefix(arg, "--sandbox=") || strings.HasPrefix(arg, "-s=") {
			allDropped = append(allDropped, arg)
			continue
		}
		kept = append(kept, arg)
	}
	return kept, allDropped
}

func filterUnsafeBypassArgs(custom, dropped []string) (kept, allDropped []string) {
	blocked := make(map[string]struct{}, len(UnsafeBypassArgs()))
	for _, arg := range UnsafeBypassArgs() {
		blocked[arg] = struct{}{}
	}

	allDropped = append(allDropped, dropped...)
	for _, arg := range custom {
		if _, isBlocked := blocked[arg]; isBlocked {
			allDropped = append(allDropped, arg)
			continue
		}
		if strings.HasPrefix(arg, "--yolo=") ||
			strings.HasPrefix(arg, "--dangerously-bypass-approvals-and-sandbox=") {
			allDropped = append(allDropped, arg)
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
func buildEnv(caller map[string]string, _ StartOptions) []string {
	reserved := map[string]string{}

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
