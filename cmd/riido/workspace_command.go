package main

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func runWorkspace(c contextIO, raw []string) error {
	requestID := contextcli.NewRequestID()
	if len(raw) == 0 || isHelpArg(raw[0]) {
		printWorkspaceUsage(c.stderr)
		if len(raw) == 0 {
			return c.fail(false, requestID, contextcli.Fail(contextcli.CodeUsage, "workspace 하위 명령이 필요합니다", "riido workspace --help"))
		}
		return nil
	}
	if raw[0] == "use" {
		return runWorkspaceUse(c, raw[1:])
	}
	if raw[0] != "list" {
		return c.fail(wantsJSON(raw), requestID, contextcli.Fail(contextcli.CodeUsage, "지원하지 않는 workspace 하위 명령입니다", "riido workspace list --help"))
	}
	args, cliErr := c.parseArgs(raw[1:], contextcli.ArgSpec{AllowLimit: true})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printWorkspaceUsage(c.stderr)
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "workspace list는 사용자 모드에서만 지원합니다", "에이전트는 배정된 작업의 범위를 사용하세요"))
	}
	if args.HasValue("--workspace") {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "workspace list는 --workspace를 지원하지 않습니다", "현재 로그인에 허용된 워크스페이스만 조회합니다"))
	}
	q := userapi.WorkspaceListRequest{Limit: 50}
	if args.HasValue("--limit") {
		value, err := strconv.Atoi(args.Value("--limit"))
		if err != nil {
			return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "--limit은 1~100의 정수여야 합니다", "riido workspace list --help"))
		}
		q.Limit = value
	}
	if args.HasValue("--cursor") {
		q.Cursor = &args.Cursor
	}
	if !q.Valid() {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "페이지 입력이 유효하지 않습니다", "--limit은 1~100, --cursor는 이전 응답의 nextCursor를 같은 limit으로 사용하세요"))
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	page, err := session.Workspaces(ctx, q)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	env := contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "live": true, "workspaces": page})
	return c.succeed(args.JSON, env, func(w io.Writer, _ any) { renderWorkspaces(w, page) })
}

func printWorkspaceUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido workspace use <WORKSPACE-ID> [--profile PROFILE] [--json]")
	fmt.Fprintln(w, "    Verify current membership, then save a profile default. API permissions are checked again on every read.")
	fmt.Fprintln(w, "  riido workspace list [--limit N] [--cursor CURSOR] [--json]")
	fmt.Fprintln(w, "    User mode only; current active memberships, one page (default 50, max 100). Continue with nextCursor and the same limit. No default workspace change.")
}

func renderWorkspaces(w io.Writer, page *userapi.WorkspaceList) {
	fmt.Fprintf(w, "현재 참여 중인 워크스페이스: 이번 페이지 %d개\n", len(page.Items))
	for _, item := range page.Items {
		fmt.Fprintf(w, "  - %q (id=%q, demo=%t)\n", item.Name, item.ID, item.IsDemo)
	}
	if page.NextCursor != nil {
		fmt.Fprintf(w, "다음 페이지: riido workspace list --limit %d --cursor %q\n", page.Limit, *page.NextCursor)
	}
}
