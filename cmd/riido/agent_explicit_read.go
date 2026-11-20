package main

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func agentExplicitWorkCall(sub workCommand) contextReadCall {
	values := []string{"--page", "--version", "--comment-id"}
	if sub == workCommandActivity {
		values = []string{"--page"}
	}
	if sub == workCommandBody {
		values = append(values, "--part", "--offset", "--comments-limit")
	}
	return contextReadCall{command: "work " + string(sub), spec: contextcli.ArgSpec{MaxPositionals: 1, Values: values, AllowLimit: true}, usage: printWorkUsage, payload: func(args contextcli.Args) (map[string]string, *contextcli.CLIError) {
		if len(args.Positionals) != 1 {
			return nil, contextcli.Fail(contextcli.CodeUsage, "작업을 지정해야 합니다", "riido work --help")
		}
		ref, err := contextcli.ParseWorkRef(args.Positionals[0])
		if err != nil {
			return nil, err
		}
		payload := map[string]string{"work": ref.Value, "work_ref_kind": string(ref.Kind)}
		for _, flag := range append(values, "--limit") {
			if args.HasValue(flag) {
				payload[strings.TrimPrefix(flag, "--")] = args.Value(flag)
			}
		}
		if sub == workCommandBody {
			if payload["version"] == "" {
				return nil, contextcli.Fail(contextcli.CodeUsage, "원 자료의 version이 필요합니다", "work get 응답의 본문 version·nextOffset을 사용하세요")
			}
			if payload["part"] == "" {
				payload["part"] = "document"
			}
		}
		return payload, nil
	}}
}

func agentThreadCall(sub string) contextReadCall {
	n := 0
	if sub == "show" {
		n = 1
	}
	return contextReadCall{command: "thread " + sub, spec: contextcli.ArgSpec{MaxPositionals: n, Values: []string{"--work"}}, usage: printThreadUsage, payload: func(args contextcli.Args) (map[string]string, *contextcli.CLIError) {
		if len(args.Positionals) != n || args.Value("--work") == "" || args.Cursor != "" {
			return nil, contextcli.Fail(contextcli.CodeUsage, "작업과 show의 대화 ID가 필요합니다. 페이지는 지원하지 않습니다", "riido thread --help")
		}
		ref, err := contextcli.ParseWorkRef(args.Value("--work"))
		if err != nil {
			return nil, err
		}
		p := map[string]string{"work": ref.Value}
		if n == 1 {
			p["thread"] = args.Positionals[0]
		}
		return p, nil
	}}
}
