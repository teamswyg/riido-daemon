package main

import (
	"strconv"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

type contextCommand string

const (
	contextCommandGet     contextCommand = "get"
	contextCommandChanges contextCommand = "changes"
)

func runContext(c contextIO, args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		printContextGroupUsage(c.stderr)
		if len(args) == 0 {
			return c.fail(false, contextcli.NewRequestID(), contextcli.Fail(contextcli.CodeUsage, "context 하위 명령이 필요합니다", "riido context --help"))
		}
		return nil
	}
	switch contextCommand(args[0]) {
	case contextCommandGet:
		return c.runContextRead(contextGetCall(), args[1:])
	case contextCommandChanges:
		return c.runContextRead(contextChangesCall(), args[1:])
	default:
		return c.fail(wantsJSON(args), contextcli.NewRequestID(),
			contextcli.Failf(contextcli.CodeUsage, "알 수 없는 context 하위 명령: %s", args[0]))
	}
}

func contextGetCall() contextReadCall {
	return contextReadCall{
		command: "context get",
		spec:    contextcli.ArgSpec{Values: []string{"--work", "--budget"}},
		usage:   printContextGroupUsage,
		payload: func(args contextcli.Args) (map[string]string, *contextcli.CLIError) {
			ref, cliErr := contextcli.ParseWorkRef(args.Value("--work"))
			if cliErr != nil {
				return nil, cliErr
			}
			payload := map[string]string{"work": ref.Value, "work_ref_kind": string(ref.Kind)}
			if budget := args.Value("--budget"); budget != "" {
				chars, err := strconv.Atoi(budget)
				if err != nil || chars < 1 {
					return nil, contextcli.Failf(contextcli.CodeUsage, "--budget 은 1 이상의 문자 수여야 합니다: %s", budget)
				}
				payload["budget_chars"] = strconv.Itoa(chars)
			}
			return payload, nil
		},
	}
}

func contextChangesCall() contextReadCall {
	return contextReadCall{
		command: "context changes",
		spec:    contextcli.ArgSpec{Values: []string{"--since"}},
		usage:   printContextGroupUsage,
		payload: func(args contextcli.Args) (map[string]string, *contextcli.CLIError) {
			since := strings.TrimSpace(args.Value("--since"))
			if since == "" {
				return nil, contextcli.Fail(contextcli.CodeUsage, "--since 가 필요합니다", "riido context get 이 돌려준 snapshot ID를 넘기세요")
			}
			return map[string]string{"since": since}, nil
		},
	}
}
