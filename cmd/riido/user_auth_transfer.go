package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/teamswyg/riido-daemon/internal/userauth"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func (c contextIO) runAuthTransfer(raw []string) error {
	id := contextcli.NewRequestID()
	args, e := c.parseArgs(raw, contextcli.ArgSpec{Bools: []string{"--repair", "--restart"}})
	if e != nil {
		return c.fail(wantsJSON(raw), id, e)
	}
	if args.Help {
		printAuthUsage(c.stderr)
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, id, contextcli.Fail(contextcli.CodePermissionDenied, "기기 전환은 사용자 터미널에서 실행해 주세요", "riido auth transfer"))
	}
	if e = checkUserOptions(args); e != nil {
		return c.fail(args.JSON, id, e)
	}
	repair, restart := args.Bool("--repair"), args.Bool("--restart")
	if restart && (!repair || args.NoInput) {
		return c.fail(args.JSON, id, contextcli.Fail(contextcli.CodeUsage, "--restart는 대화형 복구에서 사용할 수 있습니다", "riido auth transfer --repair --restart"))
	}
	if !args.HasValue("--timeout") {
		args.Timeout = 5 * time.Minute
	}
	s, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, id, userLinkError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	res, err := s.Transfer(ctx, cliLinkDaemon{io: c}, !args.NoInput, repair, restart)
	if err != nil {
		return c.fail(args.JSON, id, userTransferError(err))
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(id, res), func(w io.Writer, _ any) {
		fmt.Fprintln(w, "기기 전환 상태:", res.State)
		switch res.State {
		case "expired", "rejected":
			fmt.Fprintln(w, "새 복구 승인: riido auth transfer --repair --restart")
		case "proof_stopped":
			fmt.Fprintln(w, "기존 전환 복구: riido auth transfer --repair")
		case "awaiting_evidence", "awaiting_receipt":
			fmt.Fprintln(w, "서버 확인이 남아 있습니다. 같은 명령으로 다시 확인해 주세요.")
		}
	})
}

func userTransferError(err error) *contextcli.CLIError {
	var e *userauth.DeviceLinkError
	if errors.As(err, &e) {
		switch e.Code {
		case "TRANSFER_NO_REQUEST", "TRANSFER_NO_JOURNAL", "TRANSFER_NO_REPAIR":
			return contextcli.Fail(contextcli.CodeConflict, "재개할 기기 전환 요청을 확인하지 못했습니다", "처음 시작할 때는 riido auth transfer 를 실행하세요")
		case "TRANSFER_ACTIVE":
			return contextcli.Fail(contextcli.CodeConflict, "서버에 진행 중인 복구 요청이 있습니다", "riido auth transfer --repair")
		}
	}
	var re *userauth.InstallationRepairError
	if errors.As(err, &re) {
		return contextcli.Fail(contextcli.CodeUnavailable, "서버가 기기 전환 복구를 확인하지 못했습니다", "기존 요청으로 riido auth transfer --repair 를 다시 실행하세요")
	}
	return userLinkError(err)
}
