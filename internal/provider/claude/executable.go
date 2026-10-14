package claude

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func resolveExecutable(req agentbridge.StartRequest, opts StartOptions) string {
	switch {
	case opts.Executable != "":
		return opts.Executable
	case req.Executable != "":
		return req.Executable
	default:
		return DefaultExecutable
	}
}
