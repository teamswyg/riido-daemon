package cursor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func appendWorkspaceArgs(args []string, cwd string) []string {
	if cwd == "" {
		return args
	}
	args = append(args, "--workspace", cwd)
	return append(args, "--trust")
}

func appendModelResumeArgs(args []string, req agentbridge.StartRequest) []string {
	if req.Model != "" {
		args = append(args, "--model", req.Model)
	}
	if req.ResumeSessionID != "" {
		args = append(args, "--resume", req.ResumeSessionID)
	}
	return args
}
