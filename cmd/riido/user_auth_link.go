package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func (c contextIO) runAuthLink(raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printAuthUsage(c.stderr)
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodePermissionDenied, "에이전트 실행에서는 사용자 기기 연결을 승인할 수 없습니다", "사람의 터미널에서 riido auth link 를 실행하세요"))
	}
	if cliErr := checkUserOptions(args); cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	if !hasTimeout(raw) {
		args.Timeout = 5 * time.Minute
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userLinkError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	result, err := session.Link(ctx, cliLinkDaemon{io: c}, !args.NoInput)
	if err != nil {
		return c.fail(args.JSON, requestID, userLinkError(err))
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, result), func(w io.Writer, _ any) {
		fmt.Fprintln(w, "Riido CLI 기기 연결을 서버에서 확인했습니다.")
	})
}

var _ standalonecli.LinkDaemon = cliLinkDaemon{}
