package cursor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func BuildStart(req agentbridge.StartRequest, opts StartOptions) (agentbridge.StartCommand, error) {
	args, err := profileArgs(startProfile(opts), req.Prompt)
	if err != nil {
		return agentbridge.StartCommand{}, err
	}
	args, err = appendUnsafeBypass(args, opts)
	if err != nil {
		return agentbridge.StartCommand{}, err
	}
	args = appendWorkspaceArgs(args, req.Cwd)
	args = appendModelResumeArgs(args, req)

	kept, dropped := agentbridge.FilterBlockedArgs(req.CustomArgs, BlockedArgs())
	args = append(args, kept...)
	dropped = appendUnsupported(dropped, req)

	return agentbridge.StartCommand{
		Executable:  startExecutable(req, opts),
		Args:        args,
		Env:         envList(req.Env),
		Dir:         req.Cwd,
		StdinMode:   agentbridge.StdinNone,
		DroppedArgs: dropped,
	}, nil
}
