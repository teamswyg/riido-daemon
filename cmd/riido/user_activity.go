package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (c contextIO) runUserActivity(raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{MaxPositionals: 1, Values: []string{"--team-id", "--page"}, AllowLimit: true})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printWorkUsage(c.stderr)
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "에이전트 모드는 work activity를 지원하지 않습니다", "사용자 터미널에서 작업 ID와 --team-id로 조회하세요"))
	}
	if len(args.Positionals) != 1 || !args.HasValue("--team-id") {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "작업 ID와 --team-id가 필요합니다", "riido work activity <ID> --team-id <TEAM-ID>"))
	}
	if args.HasValue("--workspace") || args.HasValue("--cursor") {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "사용자 활동 조회는 workspace나 cursor를 지원하지 않습니다", "--page와 --limit을 사용하세요"))
	}
	work, cliErr := userActivityID(args.Positionals[0])
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	team, cliErr := userActivityID(args.Value("--team-id"))
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	pagination := userapi.ActivityPagination{Page: "1", Limit: "20"}
	if args.HasValue("--page") {
		pagination.Page = args.Value("--page")
	}
	if args.HasValue("--limit") {
		pagination.Limit = args.Value("--limit")
	}
	if !pagination.Valid() {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "page는 양의 int32 정수, limit은 1~100이며 offset은 int32 범위여야 합니다", "부호나 공백 없는 십진 정수를 사용하세요"))
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	page, err := session.Activity(ctx, work, team, pagination)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "activityLogs": page.ActivityLogs, "meta": page.Meta}), nil)
}

func userActivityID(raw string) (string, *contextcli.CLIError) {
	ref, cliErr := contextcli.ParseWorkRef(raw)
	if cliErr != nil {
		return "", cliErr
	}
	if ref.Kind != contextcli.WorkRefID {
		return "", contextcli.Fail(contextcli.CodeUnsupported, "사용자 활동 조회에는 표시 key 대신 실제 ID가 필요합니다", "작업과 팀의 실제 ID를 지정하세요")
	}
	return ref.Value, nil
}
