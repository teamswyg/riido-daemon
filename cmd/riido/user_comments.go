package main

import (
	"context"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (c contextIO) runUserComments(sub workCommand, raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{MaxPositionals: 1, Values: []string{"--team-id", "--comment-id", "--page"}, AllowLimit: sub == workCommandComments})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printWorkUsage(c.stderr)
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "에이전트 모드는 work replies를 지원하지 않습니다", "riido work comments <작업> --cursor <다음 커서>를 사용하세요"))
	}
	work, team, parent, pagination, cliErr := userCommentArguments(sub, args)
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	data := map[string]any{"mode": "user"}
	if sub == workCommandComments {
		page, err := session.Comments(ctx, work, team, pagination)
		if err != nil {
			return c.fail(args.JSON, requestID, userAuthError(err))
		}
		data["comments"], data["meta"] = page.Comments, page.Meta
	} else {
		replies, err := session.Replies(ctx, work, team, parent)
		if err != nil {
			return c.fail(args.JSON, requestID, userAuthError(err))
		}
		data["replies"] = replies
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, data), nil)
}

func userCommentArguments(sub workCommand, args contextcli.Args) (work, team, parent string, pagination userapi.CommentPagination, cliErr *contextcli.CLIError) {
	if len(args.Positionals) != 1 || strings.TrimSpace(args.Value("--team-id")) == "" {
		cliErr = contextcli.Fail(contextcli.CodeUsage, "작업 ID와 --team-id가 필요합니다", "riido work "+string(sub)+" --help")
		return work, team, parent, pagination, cliErr
	}
	if args.HasValue("--workspace") || args.HasValue("--cursor") || (sub == workCommandReplies && args.HasValue("--page")) || (sub == workCommandComments && args.HasValue("--comment-id")) {
		cliErr = contextcli.Fail(contextcli.CodeUnsupported, "이 사용자 조회에서 지원하지 않는 옵션입니다", "댓글 목록은 --page와 --limit, 답글은 --comment-id를 사용하세요")
		return work, team, parent, pagination, cliErr
	}
	if work, cliErr = userCommentID(args.Positionals[0]); cliErr != nil {
		return work, team, parent, pagination, cliErr
	}
	if team, cliErr = userCommentID(args.Value("--team-id")); cliErr != nil {
		return work, team, parent, pagination, cliErr
	}
	if sub == workCommandReplies {
		parent, cliErr = userCommentID(args.Value("--comment-id"))
		return work, team, parent, pagination, cliErr
	}
	pagination = userapi.CommentPagination{Page: "1", Limit: "20"}
	if args.HasValue("--page") {
		pagination.Page = args.Value("--page")
	}
	if args.HasValue("--limit") {
		pagination.Limit = args.Value("--limit")
	}
	if !pagination.Valid() {
		cliErr = contextcli.Fail(contextcli.CodeUsage, "페이지는 양의 int32 정수, limit은 1~100이며 offset은 int32 범위여야 합니다", "riido work comments <ID> --team-id <TEAM-ID> --page 1 --limit 20")
	}
	return work, team, parent, pagination, cliErr
}

func userCommentID(raw string) (string, *contextcli.CLIError) {
	ref, cliErr := contextcli.ParseWorkRef(raw)
	if cliErr != nil {
		return "", cliErr
	}
	if ref.Kind != contextcli.WorkRefID {
		return "", contextcli.Fail(contextcli.CodeUnsupported, "현재 사용자 댓글 조회는 표시 key 대신 ID를 사용합니다", "작업, 팀, 댓글의 실제 ID를 지정하세요")
	}
	return ref.Value, nil
}
