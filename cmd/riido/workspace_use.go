package main

import (
	"context"
	"fmt"
	"io"

	"github.com/teamswyg/riido-daemon/internal/cliconfig"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func runWorkspaceUse(c contextIO, raw []string) error {
	requestID := contextcli.NewRequestID()
	args, settings, cliErr := c.parseSettings(raw, contextcli.ArgSpec{MaxPositionals: 1})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printWorkspaceUsage(c.stderr)
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "workspace use는 사용자 모드에서만 지원합니다", "에이전트 실행의 워크스페이스는 고정됩니다"))
	}
	if len(args.Positionals) != 1 || args.HasValue("--workspace") || args.HasValue("--cursor") {
		return c.fail(args.JSON, requestID, settingsError(cliconfig.ErrInvalid))
	}
	return c.selectWorkspace(requestID, args, settings, args.Positionals[0])
}

func (c contextIO) selectWorkspace(requestID string, args contextcli.Args, settings *commandSettings, id string) error {
	if settings.resolved.AuthOrigin != settings.resolved.Stored.AuthenticationOrigin() {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeConflict, "일회용 로그인 선택으로 기본 워크스페이스를 저장할 수 없습니다", "먼저 riido config set auth_origin 으로 사용할 로그인 방식을 선택하세요"))
	}

	if !cliconfig.ValidID(id) {
		return c.fail(args.JSON, requestID, settingsError(cliconfig.ErrInvalid))
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	item, err := session.Workspace(ctx, id)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	if err = settings.store.Set(ctx, settings.resolved, "workspace", item.ID); err != nil {
		return c.fail(args.JSON, requestID, settingsError(err))
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, map[string]any{"saved": true, "profile": settings.resolved.Profile, "workspace": item.ID}), func(w io.Writer, _ any) {
		fmt.Fprintf(w, "프로필 %q의 기본 워크스페이스: %q (%q)\n", settings.resolved.Profile, item.Name, item.ID)
	})
}
