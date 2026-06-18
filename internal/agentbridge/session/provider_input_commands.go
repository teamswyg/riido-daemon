package session

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (e commandExecutor) writeProviderInput(input []byte) []agentbridge.Event {
	if len(input) == 0 {
		return nil
	}
	if err := e.proc.WriteStdin(input); err != nil {
		return []agentbridge.Event{{
			Kind: agentbridge.EventWarning,
			Text: "provider input write failed",
			Err:  err.Error(),
		}}
	}
	return nil
}

func (e commandExecutor) writeApprovalCommand(cmd agentbridge.Command) []agentbridge.Event {
	builder, ok := e.adapter.(agentbridge.ProviderInputBuilder)
	if !ok {
		return []agentbridge.Event{{
			Kind: agentbridge.EventWarning,
			Text: "provider approval command has no input builder",
		}}
	}
	input, err := builder.BuildProviderInput(cmd)
	if err != nil {
		return []agentbridge.Event{{
			Kind: agentbridge.EventWarning,
			Text: "provider approval command build failed",
			Err:  err.Error(),
		}}
	}
	if len(input) == 0 {
		return nil
	}
	if err := e.proc.WriteStdin(input); err != nil {
		return []agentbridge.Event{{
			Kind: agentbridge.EventWarning,
			Text: "provider approval command write failed",
			Err:  err.Error(),
		}}
	}
	return nil
}
