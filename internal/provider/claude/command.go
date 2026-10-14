package claude

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// BuildStart turns an agentbridge.StartRequest + Claude-specific options
// into a StartCommand. Custom args from req are filtered against
// BlockedArgs; dropped args land in StartCommand.DroppedArgs so the
// session actor can emit a Warning event per spec §9.1.
func BuildStart(req agentbridge.StartRequest, opts StartOptions) (agentbridge.StartCommand, error) {
	if err := validateStartPermission(opts); err != nil {
		return agentbridge.StartCommand{}, err
	}
	args, dropped, tempFiles := buildStartArgs(req, opts)
	return agentbridge.StartCommand{
		Executable:  resolveExecutable(req, opts),
		Args:        args,
		Env:         buildStartEnv(req),
		Dir:         req.Cwd,
		StdinMode:   agentbridge.StdinPipe,
		DroppedArgs: dropped,
		TempFiles:   tempFiles,
	}, nil
}
