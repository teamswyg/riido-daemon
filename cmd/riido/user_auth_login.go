package main

import (
	"context"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func (c contextIO) runAuthLogin(raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{Bools: []string{"--desktop", "--browser", "--with-token"}})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printAuthUsage(c.stderr)
		return nil
	}
	chosen := 0
	for _, flag := range []string{"--desktop", "--browser", "--with-token"} {
		if args.Bool(flag) {
			chosen++
		}
	}
	if chosen > 1 {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "로그인 방법은 하나만 고르세요", "riido auth login --browser"))
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodePermissionDenied, "에이전트 실행에서는 사용자 로그인을 할 수 없습니다", "에이전트는 실행 capability로만 조회합니다"))
	}
	if args.Bool("--desktop") || args.Bool("--with-token") {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "이 로그인 방법은 아직 지원하지 않습니다", "riido auth login --browser"))
	}
	if args.NoInput {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeAuthRequired, "로그인에는 브라우저와 OS 사용자 확인이 필요합니다", "대화형 터미널에서 riido auth login 을 실행하세요"))
	}
	if cliErr := checkUserOptions(args); cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	if !hasTimeout(raw) {
		args.Timeout = 5 * time.Minute
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	result, err := session.Login(ctx)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, result), nil)
}

func hasTimeout(raw []string) bool {
	for _, arg := range raw {
		if arg == "--timeout" || strings.HasPrefix(arg, "--timeout=") {
			return true
		}
	}
	return false
}

func checkUserOptions(args contextcli.Args) *contextcli.CLIError {
	if args.Workspace != "" || args.Cursor != "" {
		return contextcli.Fail(contextcli.CodeUnsupported, "현재 사용자 조회에는 --workspace 또는 --cursor를 지정할 수 없습니다", "로그인 시 승인한 범위에서 작업 ID와 --team-id를 사용하세요")
	}
	return nil
}
