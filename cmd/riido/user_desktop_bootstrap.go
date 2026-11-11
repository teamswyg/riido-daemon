package main

import (
	"context"
	"io"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// A private Desktop entry point. No request/credential material is accepted in argv.
func (c contextIO) runDesktopBootstrap(raw []string) error {
	id := contextcli.NewRequestID()
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(true, id, contextcli.Fail(contextcli.CodePermissionDenied, "기기 전환은 사용자 Desktop에서 시작해 주세요", ""))
	}
	if len(raw) == 0 || raw[0] != "identity" && raw[0] != "request" && raw[0] != "restart" {
		return c.fail(true, id, contextcli.Fail(contextcli.CodeUsage, "기기 전환 작업을 확인해 주세요", ""))
	}
	args, e := c.parseArgs(raw[1:], contextcli.ArgSpec{})
	if e != nil {
		return c.fail(true, id, e)
	}
	if e = checkUserOptions(args); e != nil {
		return c.fail(true, id, e)
	}
	if args.Help || !args.JSON || len(args.Positionals) != 0 {
		return c.fail(true, id, contextcli.Fail(contextcli.CodeUsage, "기기 전환에는 Desktop JSON 모드가 필요합니다", ""))
	}
	var payload []byte
	if raw[0] != "identity" {
		if c.stdin == nil {
			return c.fail(true, id, contextcli.Fail(contextcli.CodeUsage, "기기 전환 요청을 확인해 주세요", ""))
		}
		var err error
		payload, err = io.ReadAll(io.LimitReader(c.stdin, 4097))
		if err != nil || len(payload) == 0 || len(payload) > 4096 {
			return c.fail(true, id, contextcli.Fail(contextcli.CodeUsage, "기기 전환 요청을 확인해 주세요", ""))
		}
	}
	s, err := c.openUserSession(args)
	if err != nil {
		return c.fail(true, id, userLinkError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	out, err := s.DesktopBootstrap(ctx, raw[0], payload)
	if err != nil {
		return c.fail(true, id, userTransferError(err))
	}
	return c.succeed(true, contextcli.NewEnvelope(id, out), nil)
}
