package cursor

import (
	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
)

const (
	Name              = string(providercatalog.KindCursor)
	DefaultExecutable = "cursor-agent"
	EnvOverride       = "RIIDO_CURSOR_PATH"
	APIKeyEnv         = "CURSOR_API_KEY"
	MaxLineBytes      = 10 * 1024 * 1024
)

func BlockedArgs() []string {
	return providercap.ProtocolCriticalArgs(providercap.ProtocolCursorAgentStreamJSON)
}
