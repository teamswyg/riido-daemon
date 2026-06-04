// Package codex owns the C4 run-scope adapter for OpenAI's Codex CLI. The
// spawn shape is `codex app-server --listen stdio://` so the session actor
// speaks JSON-RPC 2.0 over the process's stdio pipes.
//
// What this slice provides:
//   - Command builder with protocol-critical args locked in.
//   - --listen blocklist (caller cannot reroute transport).
//   - task-scoped Codex permission profile injection.
//   - JSON-RPC protocol driver via the provider-neutral agentbridge port.
package codex

import (
	"fmt"
	"maps"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const Name = "codex"
const DefaultExecutable = "codex"
const DefaultPermissionProfile = "riido-task"

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

// SecurityCriticalArgs are Codex app-server flags that can rewrite the
// daemon-owned permission profile. They are distinct from protocol-critical
// args: --listen protects transport shape, while these protect C7 sandbox
// policy injection.
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
	// AuthHomeDenyPath is the user-global Codex credential/config home path
	// that provider tool commands must not read. The Codex app-server process
	// may still use its inherited auth store, but every spawned shell command
	// receives a daemon-owned permission profile with this path set to none.
	AuthHomeDenyPath string
	// PermissionProfile overrides the generated Codex permission profile name.
	// Empty means DefaultPermissionProfile.
	PermissionProfile string
}

type permissionEntry struct {
	path   string
	access string
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
		"app-server",
	}
	args = append(args, permissionProfileArgs(req, opts)...)
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

func filterCustomArgs(custom []string) (kept []string, dropped []string) {
	kept, dropped = agentbridge.FilterBlockedArgs(custom, BlockedArgs())
	kept, dropped = filterConfigOverrideArgs(kept, dropped)
	return filterUnsafeBypassArgs(kept, dropped)
}

func filterConfigOverrideArgs(custom []string, dropped []string) (kept []string, allDropped []string) {
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

func permissionProfileArgs(req agentbridge.StartRequest, opts StartOptions) []string {
	profile := strings.TrimSpace(opts.PermissionProfile)
	if profile == "" {
		profile = DefaultPermissionProfile
	}
	cwd := cleanAbs(req.Cwd)
	authHome := cleanAbs(firstNonEmpty(
		opts.AuthHomeDenyPath,
		startEnvValue(req.Env, "CODEX_HOME"),
		defaultCodexHomeFromEnv(req.Env),
	))

	entries := []permissionEntry{}
	addEntry := func(path string, access string) {
		path = cleanAbs(path)
		if path == "" {
			return
		}
		for i := range entries {
			if entries[i].path == path {
				entries[i].access = access
				return
			}
		}
		entries = append(entries, permissionEntry{path: path, access: access})
	}

	addEntry(":minimal", "read")
	if cwd != "" {
		addEntry(cwd, "write")
	}
	for _, path := range commonToolchainReadPaths(req.Env) {
		addEntry(path, "read")
	}
	for _, path := range commonToolchainWritePaths(req.Env) {
		addEntry(path, "write")
	}
	if authHome != "" {
		addEntry(authHome, "none")
	}
	inlineEntries := make([]string, 0, len(entries))
	for _, entry := range entries {
		inlineEntries = append(inlineEntries, tomlInlineEntry(entry.path, entry.access))
	}

	return []string{
		"-c", fmt.Sprintf("default_permissions=%s", strconv.Quote(profile)),
		"-c", fmt.Sprintf("permissions.%s.filesystem={%s}", profile, strings.Join(inlineEntries, ",")),
		"-c", fmt.Sprintf("permissions.%s.network={enabled=true}", profile),
	}
}

func commonToolchainReadPaths(env map[string]string) []string {
	home := startEnvValue(env, "HOME")
	paths := []string{
		startEnvValue(env, "GOROOT"),
		startEnvValue(env, "RUSTUP_HOME"),
		startEnvValue(env, "CARGO_HOME"),
	}
	if home != "" {
		paths = append(paths,
			filepath.Join(home, ".rustup"),
			filepath.Join(home, ".cargo"),
		)
	}
	if home != "" || startEnvValue(env, "GOROOT") != "" {
		paths = append(paths,
			"/usr/local/go",
			"/opt/homebrew/opt/go",
			"/usr/local/opt/go",
		)
	}
	return paths
}

func commonToolchainWritePaths(env map[string]string) []string {
	home := startEnvValue(env, "HOME")
	paths := []string{
		startEnvValue(env, "GOCACHE"),
	}
	if home != "" {
		paths = append(paths,
			filepath.Join(home, "Library", "Caches", "go-build"),
			filepath.Join(home, ".cache", "go-build"),
		)
	}
	return paths
}

func tomlInlineEntry(path string, access string) string {
	return strconv.Quote(path) + "=" + strconv.Quote(access)
}

func cleanAbs(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !filepath.IsAbs(path) {
		return path
	}
	return filepath.Clean(path)
}

func startEnvValue(env map[string]string, key string) string {
	if env == nil {
		return ""
	}
	return strings.TrimSpace(env[key])
}

func defaultCodexHomeFromEnv(env map[string]string) string {
	home := startEnvValue(env, "HOME")
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".codex")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
