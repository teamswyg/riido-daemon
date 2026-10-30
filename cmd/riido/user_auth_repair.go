package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func (c contextIO) runAuthRepair(raw []string) error {
	id := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{Bools: []string{"--restart"}})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), id, cliErr)
	}
	if args.Help {
		printAuthUsage(c.stderr)
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, id, contextcli.Fail(contextcli.CodePermissionDenied, "에이전트 실행에서는 사용자 연결을 복구할 수 없습니다", "사람의 터미널에서 riido auth repair 를 실행하세요"))
	}
	if cliErr := checkUserOptions(args); cliErr != nil {
		return c.fail(args.JSON, id, cliErr)
	}
	if args.NoInput && args.Bool("--restart") {
		return c.fail(args.JSON, id, contextcli.Fail(contextcli.CodeUsage, "새 복구 요청은 사용자 확인이 필요합니다", "--no-input 없이 다시 실행하세요"))
	}
	if !hasTimeout(raw) {
		args.Timeout = 5 * time.Minute
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, id, userRepairError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	result, err := session.Repair(ctx, cliLinkDaemon{io: c}, !args.NoInput, args.Bool("--restart"))
	if err != nil {
		return c.fail(args.JSON, id, userRepairError(err))
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(id, result), func(w io.Writer, _ any) {
		fmt.Fprintln(w, "Riido CLI 기기 연결 복구를 서버에서 확인했습니다. 읽기 권한이 필요하면 riido auth login 을 실행하세요.")
	})
}
