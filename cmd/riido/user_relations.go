package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func (c contextIO) runUserRelations(raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := c.parseUserWorkArgs(raw, contextcli.ArgSpec{MaxPositionals: 1, Values: []string{"--team-id", "--page"}}, true)
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printWorkUsage(c.stderr)
		return nil
	}
	if len(args.Positionals) != 1 {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "작업 ID와 --team-id가 필요합니다", "riido work relations <ID> --team-id <TEAM-ID>"))
	}
	if args.HasValue("--cursor") || args.HasValue("--page") {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "사용자 관계 조회는 페이지 옵션을 지원하지 않습니다", "작업 ID와 팀 ID 또는 표시 키와 워크스페이스/팀 범위를 지정하세요"))
	}
	work, team, cliErr := userWorkTarget(args, args.Positionals[0])
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	relations, err := session.Relations(ctx, work, team, args.Workspace)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "relations": relations}), relationsView(c.stdout, relations))
}

func userRelationID(raw string) (string, *contextcli.CLIError) {
	ref, cliErr := contextcli.ParseWorkRef(raw)
	if cliErr != nil {
		return "", cliErr
	}
	if ref.Kind != contextcli.WorkRefID {
		return "", contextcli.Fail(contextcli.CodeUnsupported, "현재 사용자 관계 조회는 표시 key 대신 ID를 사용합니다", "작업과 팀의 실제 ID를 지정하세요")
	}
	return ref.Value, nil
}
