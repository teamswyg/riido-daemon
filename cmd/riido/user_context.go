package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func (c contextIO) runUserContext(raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := c.parseUserWorkArgs(raw, contextcli.ArgSpec{Values: []string{"--work", "--team-id"}}, true)
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printContextGroupUsage(c.stderr)
		return nil
	}
	work, team, cliErr := userContextArguments(args)
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	scope, err := session.Context(ctx, work, team, args.Workspace)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	env := contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "live": true, "context": scope})
	if scope.Status == "partial" {
		env.Warnings = append(env.Warnings, contextcli.Warning{Code: "CONTEXT_INCOMPLETE", Message: "일부 자료를 완전히 읽지 못했습니다. 각 자료의 상태와 다음 페이지를 확인하세요"})
	}
	if scopeHasDeferredContent(scope) {
		env.Warnings = append(env.Warnings, deferredContentWarning())
	}
	return c.succeed(args.JSON, env, contextView(c.stdout, scope))
}

func userContextArguments(args contextcli.Args) (string, string, *contextcli.CLIError) {
	if args.HasValue("--cursor") {
		return "", "", contextcli.Fail(contextcli.CodeUnsupported, "사용자 맥락 조회에 스냅샷 cursor를 사용할 수 없습니다", "작업 키나 ID를 지정하세요")
	}
	return userWorkTarget(args, args.Value("--work"))
}
