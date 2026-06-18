package claude

import providercap "github.com/teamswyg/riido-contracts/provider/capability"

// BlockedArgs lists the protocol-critical flags this adapter manages itself.
// Custom args containing these are dropped with a Warning event.
func BlockedArgs() []string {
	return providercap.ProtocolCriticalArgs(providercap.ProtocolClaudeStreamJSON)
}
