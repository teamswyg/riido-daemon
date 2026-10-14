package codex

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
