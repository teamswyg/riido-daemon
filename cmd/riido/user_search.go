package main

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (c contextIO) runUserSearch(raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{Values: []string{"--query", "--page", "--budget"}, AllowLimit: true})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printWorkUsage(c.stderr)
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "work search는 사용자 모드에서만 지원합니다", "사용자 터미널에서 워크스페이스를 지정하세요"))
	}
	if args.HasValue("--page") || args.HasValue("--budget") {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "작업 검색은 page/budget을 지원하지 않습니다", "nextCursor를 --cursor로 이어 읽으세요"))
	}
	q := userapi.WorkSearchRequest{WorkspaceID: args.Workspace, Query: args.Value("--query"), Cursor: args.Cursor, Limit: 20}
	if args.HasValue("--limit") {
		limit, err := strconv.Atoi(args.Value("--limit"))
		if err != nil || strconv.Itoa(limit) != args.Value("--limit") || limit < 1 || limit > 100 {
			return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "limit은 1~100 정수여야 합니다", "--limit 20"))
		}
		q.Limit = limit
	}
	if !args.HasValue("--workspace") || !q.Valid() {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "워크스페이스 ID와 1~200자 검색어가 필요합니다", "riido work search --workspace <WORKSPACE-ID> --query <TEXT>"))
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	result, err := session.Search(ctx, q)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	env := contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "live": true, "search": result})
	return c.succeed(args.JSON, env, func(w io.Writer, _ any) {
		fmt.Fprintf(w, "워크스페이스 %q 검색 %q\n", result.WorkspaceID, result.Query)
		for _, item := range result.Items {
			fmt.Fprintf(w, "  %q %q (id=%q, team=%q, matched=%q)\n", item.Kind, *item.Title, item.ID, item.TeamID, item.MatchedFields)
		}
		if result.Complete && len(result.Items) == 0 {
			fmt.Fprintln(w, "  검색 결과가 없습니다.")
		}
		if result.NextCursor != nil {
			fmt.Fprintf(w, "다음 페이지: 같은 --workspace, --query, --limit 값을 유지하고 --cursor %q 를 추가하세요.\n", *result.NextCursor)
		}
	})
}
