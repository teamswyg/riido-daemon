package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func (c contextIO) runUserWhoami(raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		fmt.Fprintln(c.stderr, "  riido whoami [--json]")
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
	identity, err := session.Whoami(ctx)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "identity": identity}), nil)
}

func (c contextIO) runUserWork(raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{MaxPositionals: 1, Values: []string{"--team-id"}})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printWorkUsage(c.stderr)
		return nil
	}
	if len(args.Positionals) != 1 || strings.TrimSpace(args.Value("--team-id")) == "" {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "작업 ID와 --team-id가 필요합니다", "riido work get <ID> --team-id <TEAM-ID>"))
	}
	ref, cliErr := contextcli.ParseWorkRef(args.Positionals[0])
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	if ref.Kind == contextcli.WorkRefKey {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "현재 사용자 조회는 작업 key 대신 ID를 사용합니다", "riido work get <ID> --team-id <TEAM-ID>"))
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
	work, err := session.Work(ctx, ref.Value, strings.TrimSpace(args.Value("--team-id")))
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "work": work}), nil)
}
