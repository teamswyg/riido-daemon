package cursor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func startExecutable(req agentbridge.StartRequest, opts StartOptions) string {
	if opts.Executable != "" {
		return opts.Executable
	}
	if req.Executable != "" {
		return req.Executable
	}
	return DefaultExecutable
}

func startProfile(opts StartOptions) Profile {
	if opts.Profile != "" {
		return opts.Profile
	}
	return DefaultProfile
}
