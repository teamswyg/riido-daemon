package main

import (
	"fmt"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

type workCommand string

const (
	workCommandGet       workCommand = "get"
	workCommandBody      workCommand = "body"
	workCommandTree      workCommand = "tree"
	workCommandRelations workCommand = "relations"
	workCommandComments  workCommand = "comments"
	workCommandReplies   workCommand = "replies"
	workCommandActivity  workCommand = "activity"
	workCommandSearch    workCommand = "search"
)

func runWork(c contextIO, args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		printWorkUsage(c.stderr)
		if len(args) == 0 {
			return c.fail(false, contextcli.NewRequestID(), contextcli.Fail(contextcli.CodeUsage, "work 하위 명령이 필요합니다", "riido work --help"))
		}
		return nil
	}
	sub := workCommand(args[0])
	switch sub {
	case workCommandGet, workCommandTree, workCommandRelations, workCommandComments:
		if sub == workCommandTree && c.environment().Mode != contextcli.ModeAgent {
			return c.runUserTree(args[1:])
		}
		if sub == workCommandGet && c.environment().Mode != contextcli.ModeAgent {
			return c.runUserWork(args[1:])
		}
		if sub == workCommandComments && c.environment().Mode != contextcli.ModeAgent {
			return c.runUserComments(sub, args[1:])
		}
		if sub == workCommandRelations && c.environment().Mode != contextcli.ModeAgent {
			return c.runUserRelations(args[1:])
		}
		return c.runContextRead(workRefCall(sub), args[1:])
	case workCommandActivity:
		return c.runUserActivity(args[1:])
	case workCommandReplies:
		return c.runUserComments(sub, args[1:])
	case workCommandBody:
		return c.runUserBody(args[1:])
	case workCommandSearch:
		if c.environment().Mode != contextcli.ModeAgent {
			return c.runUserSearch(args[1:])
		}
		return c.runContextRead(workSearchCall(), args[1:])
	default:
		return c.fail(wantsJSON(args), contextcli.NewRequestID(),
			contextcli.Failf(contextcli.CodeUsage, "알 수 없는 work 하위 명령: %s", args[0]))
	}
}

// workRefCall builds a read that takes exactly one work reference. Comment
// paging uses --cursor with the server's fixed page size.
func workRefCall(sub workCommand) contextReadCall {
	return contextReadCall{
		command: "work " + string(sub),
		spec:    contextcli.ArgSpec{MaxPositionals: 1},
		usage:   printWorkUsage,
		payload: func(args contextcli.Args) (map[string]string, *contextcli.CLIError) {
			if len(args.Positionals) == 0 {
				return nil, contextcli.Fail(contextcli.CodeUsage, "작업을 지정해야 합니다", fmt.Sprintf("riido work %s RIID-1234", sub))
			}
			ref, cliErr := contextcli.ParseWorkRef(args.Positionals[0])
			if cliErr != nil {
				return nil, cliErr
			}
			return map[string]string{"work": ref.Value, "work_ref_kind": string(ref.Kind)}, nil
		},
	}
}

func workSearchCall() contextReadCall {
	return contextReadCall{
		command: "work search",
		spec:    contextcli.ArgSpec{Values: []string{"--query"}},
		usage:   printWorkUsage,
		payload: func(args contextcli.Args) (map[string]string, *contextcli.CLIError) {
			query := strings.TrimSpace(args.Value("--query"))
			if query == "" {
				return nil, contextcli.Fail(contextcli.CodeUsage, "--query 가 필요합니다", "riido work search --query 검색어")
			}
			return map[string]string{"query": query}, nil
		},
	}
}
