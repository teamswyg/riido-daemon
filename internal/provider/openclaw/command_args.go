package openclaw

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func buildCommandArgs(req agentbridge.StartRequest) ([]string, []string) {
	args := []string{"agent", "exec", "--json"}
	kept, dropped := agentbridge.FilterBlockedArgs(req.CustomArgs, BlockedArgs())
	args = append(args, kept...)
	return append(args, buildMessage(req.SystemPrompt, req.Prompt)), dropped
}

func envList(env map[string]string) []string {
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}
