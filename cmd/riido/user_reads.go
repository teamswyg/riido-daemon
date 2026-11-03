package main

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func (c contextIO) runUserWhoami(raw []string) error {
	requestID := contextcli.NewRequestID()
	args, settings, cliErr := c.parseSettings(raw, contextcli.ArgSpec{})
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
	view := whoamiWorkspaces{}
	if settings != nil {
		view.defaultID = settings.resolved.Workspace
	}
	if !args.JSON {
		view.names = whoamiWorkspaceNames(ctx, session, identity.LastViewedWorkspaceID, view.defaultID)
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "identity": identity}), whoamiView(c.stdout, identity, view))
}

func (c contextIO) runUserWork(raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := c.parseUserWorkArgs(raw, contextcli.ArgSpec{MaxPositionals: 1, Values: []string{"--team-id"}}, true)
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printWorkUsage(c.stderr)
		return nil
	}
	if len(args.Positionals) != 1 {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "작업 키나 ID가 필요합니다", "riido work get --help"))
	}
	work, team, cliErr := userWorkTarget(args, args.Positionals[0])
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	if args.HasValue("--cursor") {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "work get에는 cursor를 사용할 수 없습니다", "작업 키나 ID를 지정하세요"))
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	result, err := session.Work(ctx, work, team, args.Workspace)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "work": result}), workView(c.stdout, result))
}
