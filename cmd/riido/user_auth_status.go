package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func (c contextIO) runAuthStatus(raw []string) error {
	if c.environment().Mode == contextcli.ModeAgent {
		return c.runAuthLocal(raw, "에이전트 실행은 사용자 자격을 읽지 않습니다")
	}
	return c.runUserAuthLocal(raw, false)
}

func (c contextIO) runAuthLogout(raw []string) error {
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(wantsJSON(raw), contextcli.NewRequestID(), contextcli.Fail(contextcli.CodePermissionDenied, "에이전트 실행에서는 사용자 로그아웃을 할 수 없습니다", "사용자 터미널에서 실행하세요"))
	}
	return c.runUserAuthLocal(raw, true)
}

func (c contextIO) runUserAuthLocal(raw []string, logout bool) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printAuthUsage(c.stderr)
		return nil
	}
	if cliErr := checkUserOptions(args); cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	if logout {
		if err := session.Logout(ctx); err != nil {
			return c.fail(args.JSON, requestID, userAuthError(err))
		}
		return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "user_credential": "absent", "logged_out": true}), nil)
	}
	status, err := session.Status(ctx)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, status), nil)
}
