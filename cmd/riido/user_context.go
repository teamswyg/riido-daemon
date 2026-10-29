package main

import (
	"context"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func (c contextIO) runUserContext(raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{Values: []string{"--work", "--team-id"}})
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
	scope, err := session.Context(ctx, work, team)
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
	return c.succeed(args.JSON, env, nil)
}

func userContextArguments(args contextcli.Args) (string, string, *contextcli.CLIError) {
	if strings.TrimSpace(args.Value("--work")) == "" || strings.TrimSpace(args.Value("--team-id")) == "" {
		return "", "", contextcli.Fail(contextcli.CodeUsage, "작업 ID와 --team-id가 필요합니다", "riido context get --work <ID> --team-id <TEAM-ID>")
	}
	if args.HasValue("--cursor") || args.HasValue("--workspace") {
		return "", "", contextcli.Fail(contextcli.CodeUnsupported, "사용자 맥락 조회에 실행 스냅샷 커서나 다른 workspace를 지정할 수 없습니다", "작업 ID와 실제 팀 ID를 지정하세요")
	}
	work, err := userCommentID(args.Value("--work"))
	if err != nil {
		return "", "", err
	}
	team, err := userCommentID(args.Value("--team-id"))
	return work, team, err
}
