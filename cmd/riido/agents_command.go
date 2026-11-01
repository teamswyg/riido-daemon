package main

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

type agentsCommand string

const (
	agentsCommandList agentsCommand = "list"
	agentsCommandShow agentsCommand = "show"
)

// runAgents serves current visible agents through the run broker or a dedicated
// user reader. Only the run path participates in RIID-5892 mention prechecks.
func runAgents(c contextIO, args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		printAgentsUsage(c.stderr)
		if len(args) == 0 {
			return c.fail(false, contextcli.NewRequestID(), contextcli.Fail(contextcli.CodeUsage, "agents 하위 명령이 필요합니다", "riido agents --help"))
		}
		return nil
	}
	switch agentsCommand(args[0]) {
	case agentsCommandList:
		if c.environment().Mode == contextcli.ModeUser {
			return c.runUserAgents(agentsCommandList, args[1:])
		}
		return c.runContextRead(contextReadCall{
			command: "agents list",
			usage:   printAgentsUsage,
			payload: func(contextcli.Args) (map[string]string, *contextcli.CLIError) { return map[string]string{}, nil },
		}, args[1:])
	case agentsCommandShow:
		if c.environment().Mode == contextcli.ModeUser {
			return c.runUserAgents(agentsCommandShow, args[1:])
		}
		return c.runContextRead(agentsShowCall(), args[1:])
	default:
		return c.fail(wantsJSON(args), contextcli.NewRequestID(),
			contextcli.Failf(contextcli.CodeUsage, "알 수 없는 agents 하위 명령: %s", args[0]))
	}
}

func agentsShowCall() contextReadCall {
	return contextReadCall{
		command: "agents show",
		spec:    contextcli.ArgSpec{MaxPositionals: 1},
		usage:   printAgentsUsage,
		payload: func(args contextcli.Args) (map[string]string, *contextcli.CLIError) {
			if len(args.Positionals) == 0 || strings.TrimSpace(args.Positionals[0]) == "" {
				return nil, contextcli.Fail(contextcli.CodeUsage, "Agent ID를 지정해야 합니다", "riido agents list --json 으로 ID를 확인하세요")
			}
			return map[string]string{"agent": strings.TrimSpace(args.Positionals[0])}, nil
		},
	}
}
