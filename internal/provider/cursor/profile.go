package cursor

// Profile selects which cursor-agent CLI launch shape to use.
type Profile string

const (
	// ProfileRootPrint is the current cursor-agent CLI shape (2026-05+).
	// Pass -p / --output-format at the root, no subcommand.
	ProfileRootPrint Profile = "root-print"
	// ProfileAgentSubcommand uses `cursor-agent agent -p ...` for builds
	// that require the `agent` subcommand.
	ProfileAgentSubcommand Profile = "agent-subcommand"
	// ProfileLegacyChat uses `cursor-agent chat -p ...`. Opt-in only.
	ProfileLegacyChat Profile = "legacy-chat"
)

// DefaultProfile is the launch shape used when StartOptions.Profile is empty.
const DefaultProfile = ProfileRootPrint
