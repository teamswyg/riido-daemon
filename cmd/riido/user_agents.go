package main

import (
	"context"
	"fmt"
	"io"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (c contextIO) runUserAgents(command agentsCommand, raw []string) error {
	requestID := contextcli.NewRequestID()
	arity := 0
	if command == agentsCommandShow {
		arity = 1
	}
	args, cliErr := c.parseCPWorkArgs(raw, contextcli.ArgSpec{UseWorkspace: true, MaxPositionals: arity, Values: []string{"--work"}})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printAgentsUsage(c.stderr)
		return nil
	}
	if len(args.Positionals) != arity || (arity == 1 && !userapi.ValidThreadReadID(args.Positionals[0])) || (args.HasValue("--cursor") && (arity == 1 || args.Cursor == "" || !userapi.ValidAgentCursor(args.Cursor))) {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "워크스페이스·작업 ID와 올바른 목록 cursor 또는 Agent ID가 필요합니다", "riido agents --help"))
	}
	work, team, cliErr := userCPWorkTarget(args)
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	var result any
	if command == agentsCommandList {
		result, err = session.Agents(ctx, args.Workspace, work, args.Cursor, team)
	} else {
		result, err = session.Agent(ctx, args.Workspace, work, args.Positionals[0], team)
	}
	if err != nil {
		return c.fail(args.JSON, requestID, userCPReadError(err))
	}
	env := contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "live": true, "directory": result})
	env.Warnings = append(env.Warnings, contextcli.Warning{Code: "EXECUTION_POLICY_NOT_CHECKED", Message: "현재 연결 상태만 조회했습니다. 멘션·실행 허용 여부는 실행 시 서버가 판정합니다"})
	return c.succeed(args.JSON, env, func(w io.Writer, _ any) { renderUserAgents(w, result) })
}

func renderUserAgents(w io.Writer, result any) {
	if term := humanTerminal(w); term.Interactive {
		renderUserAgentsTerminal(w, term, result)
		return
	}
	fmt.Fprintln(w, "현재 볼 수 있는 Agent입니다. 연결 상태는 멘션·실행 허가를 뜻하지 않습니다.")
	printAgent := func(agent userapi.PublicAgent) {
		fmt.Fprintf(w, "  %q (%q) 역할=%q 작업=%q 연결=%q 런타임=%q 설명=%q\n", agent.Name, agent.AgentID, agent.RoleStatus, agent.WorkStatus, agent.Runtime.Connection, agent.Runtime.Kind, agent.Description)
	}
	switch value := result.(type) {
	case *userapi.AgentPage:
		fmt.Fprintf(w, "현재 조회 %d개 / 볼 수 있는 전체 %d개 (cursor=%q)\n", len(value.Agents), value.Total, value.Cursor)
		for _, agent := range value.Agents {
			printAgent(agent)
		}
		if value.NextCursor != "" {
			fmt.Fprintf(w, "다음 페이지: 같은 작업·워크스페이스로 --cursor %q\n", value.NextCursor)
		}
	case *userapi.AgentDetail:
		printAgent(value.Agent)
	}
}
