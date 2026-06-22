package openclaw

import (
	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const (
	Name              = string(providercatalog.KindOpenClaw)
	DefaultExecutable = "openclaw"
)

// BlockedArgs lists the protocol-critical flags this adapter sets itself.
// Callers cannot override these via CustomArgs.
func BlockedArgs() []string {
	return providercap.ProtocolCriticalArgs(providercap.ProtocolOpenClawAgentJSON)
}

type StartOptions struct {
	// Executable overrides the binary path.
	Executable string
	// SessionID overrides the provider-neutral session id resolution.
	// OpenClaw's resume model is session-id-based; silently using an
	// empty session id would create an anonymous run.
	SessionID string
}

func BuildStart(req agentbridge.StartRequest, opts StartOptions) (agentbridge.StartCommand, error) {
	sessionID := opts.SessionID
	if sessionID == "" {
		var err error
		sessionID, err = ResolveSessionID(req)
		if err != nil {
			return agentbridge.StartCommand{}, err
		}
	}
	exe := opts.Executable
	if exe == "" {
		exe = req.Executable
	}
	if exe == "" {
		exe = DefaultExecutable
	}

	args, dropped := buildCommandArgs(req, sessionID)
	env, tempFiles, err := buildStartEnv(req)
	if err != nil {
		return agentbridge.StartCommand{}, err
	}

	return agentbridge.StartCommand{
		Executable:  exe,
		Args:        args,
		Env:         env,
		Dir:         req.Cwd,
		StdinMode:   agentbridge.StdinNone,
		DroppedArgs: dropped,
		TempFiles:   tempFiles,
	}, nil
}
