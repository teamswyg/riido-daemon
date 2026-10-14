package codex

// StartOptions carries Codex-specific knobs.
type StartOptions struct {
	// Executable overrides the binary path. Falls back to DefaultExecutable.
	Executable string
}
