package main

import (
	"strconv"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

type contextCommand string

const (
	contextCommandGet     contextCommand = "get"
	contextCommandPart    contextCommand = "part"
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
		if c.environment().Mode != contextcli.ModeAgent {
			return c.runUserContext(args[1:])
		}
		return c.runContextRead(contextGetCall(), args[1:])
	case contextCommandPart:
		return c.runContextRead(contextPartCall(), args[1:])
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
			// Without --work an agent run reads its own assigned task.
			payload := map[string]string{"work": ""}
			if raw := args.Value("--work"); raw != "" {
				ref, cliErr := contextcli.ParseWorkRef(raw)
				if cliErr != nil {
					return nil, cliErr
				}
				payload = map[string]string{"work": ref.Value, "work_ref_kind": string(ref.Kind)}
			}
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
		spec:    contextcli.ArgSpec{Values: []string{"--since", "--budget"}},
		usage:   printContextGroupUsage,
		payload: func(args contextcli.Args) (map[string]string, *contextcli.CLIError) {
			since, cliErr := contextcli.SnapshotRef(args.Value("--since"))
			if cliErr != nil {
				return nil, cliErr
			}
			if args.HasValue("--cursor") {
				return nil, contextcli.Fail(contextcli.CodeUsage, "context changes에는 --cursor를 사용할 수 없습니다", "riido context changes --since SNAPSHOT")
			}
			payload := map[string]string{"since": since}
			if args.HasValue("--budget") {
				n, err := strconv.Atoi(args.Value("--budget"))
				if err != nil || n < 1 {
					return nil, contextcli.Fail(contextcli.CodeUsage, "--budget 은 1 이상의 문자 수여야 합니다", "riido context changes --help")
				}
				payload["budget_chars"] = strconv.Itoa(n)
			}
			return payload, nil
		},
	}
}
