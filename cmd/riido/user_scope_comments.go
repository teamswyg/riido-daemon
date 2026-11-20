package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func hasScopeCommentOptions(raw []string) bool {
	for _, arg := range raw {
		if arg == "--kind" || arg == "--version" || arg == "--locator-version" || strings.HasPrefix(arg, "--kind=") || strings.HasPrefix(arg, "--version=") || strings.HasPrefix(arg, "--locator-version=") {
			return true
		}
	}
	return false
}

func (c contextIO) runUserScopeComments(sub workCommand, raw []string) error {
	id := contextcli.NewRequestID()
	args, cliErr := c.parseUserWorkArgs(raw, contextcli.ArgSpec{MaxPositionals: 1, Values: []string{"--team-id", "--kind", "--comment-id", "--page", "--version", "--locator-version"}, AllowLimit: true}, true)
	if cliErr != nil {
		return c.fail(wantsJSON(raw), id, cliErr)
	}
	if args.Help {
		printWorkUsage(c.stderr)
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, id, contextcli.Fail(contextcli.CodeUnsupported, "에이전트는 저장된 댓글 페이지를 조회해야 합니다", "riido work comments <ID> --cursor <CURSOR> 또는 riido context part를 사용하세요"))
	}
	q, cliErr := scopeCommentArguments(sub, args)
	if cliErr != nil {
		return c.fail(args.JSON, id, cliErr)
	}
	return c.readUserScopeComments(id, args, q)
}

func (c contextIO) readUserScopeComments(id string, args contextcli.Args, q userapi.ReadScopeCommentsRequest) error {
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, id, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	if version, _ := userLocatorVersion(args); version == 2 {
		page, err := session.ScopeCommentsV2(ctx, userV2Reference(args, q.ResourceID, q.TeamID), q)
		if err != nil {
			return c.fail(args.JSON, id, userAuthError(err))
		}
		env := contextcli.NewEnvelope(id, map[string]any{"mode": "user", "live": true, "comment_page": page})
		if page.ContentPage != nil {
			env.Warnings = append(env.Warnings, deferredContentWarning())
		}
		return c.succeed(args.JSON, env, scopeCommentsV2View(c.stdout, page))
	}
	page, err := session.ScopeComments(ctx, q, args.Workspace)
	if err != nil {
		return c.fail(args.JSON, id, userAuthError(err))
	}
	env := contextcli.NewEnvelope(id, map[string]any{"mode": "user", "live": true, "comment_page": page})
	if page.ContentPage != nil {
		env.Warnings = append(env.Warnings, deferredContentWarning())
	}
	return c.succeed(args.JSON, env, scopeCommentsView(c.stdout, page))
}

func scopeCommentArguments(sub workCommand, args contextcli.Args) (userapi.ReadScopeCommentsRequest, *contextcli.CLIError) {
	q := userapi.ReadScopeCommentsRequest{ResourceKind: "component", Page: 1, Limit: 20, Version: args.Value("--version")}
	bad := func() (userapi.ReadScopeCommentsRequest, *contextcli.CLIError) {
		return userapi.ReadScopeCommentsRequest{}, contextcli.Fail(contextcli.CodeUsage, "작업 종류·ID·팀과 올바른 댓글 페이지가 필요합니다. 다음 페이지에는 원 version이 필요합니다", "riido work comments|replies --help에서 --kind, --page, --version을 확인하세요")
	}
	if len(args.Positionals) != 1 {
		return bad()
	}
	if args.HasValue("--cursor") {
		return userapi.ReadScopeCommentsRequest{}, contextcli.Fail(contextcli.CodeUnsupported, "사용자 댓글 조회는 팀 ID와 version·page를 사용합니다", "실행 스냅샷 cursor는 사용할 수 없습니다")
	}
	var cliErr *contextcli.CLIError
	if args.HasValue("--kind") {
		q.ResourceKind = args.Value("--kind")
	}
	version, versionErr := userLocatorVersion(args)
	if versionErr != nil {
		return q, versionErr
	}
	if q.ResourceKind == "component" || version == 2 {
		q.ResourceID, q.TeamID, cliErr = userWorkTarget(args, args.Positionals[0])
	} else {
		if args.HasValue("--workspace") {
			return q, contextcli.Fail(contextcli.CodeUnsupported, "이 자료에는 --workspace를 사용할 수 없습니다", "자료 ID와 팀 ID를 지정하세요")
		}
		q.ResourceID, cliErr = userCommentID(args.Positionals[0])
		if cliErr == nil {
			q.TeamID, cliErr = userCommentID(args.Value("--team-id"))
		}
	}
	if cliErr != nil {
		return q, cliErr
	}
	if sub == workCommandReplies {
		if q.ParentCommentID, cliErr = userCommentID(args.Value("--comment-id")); cliErr != nil {
			return q, cliErr
		}
	} else if args.HasValue("--comment-id") {
		return bad()
	}
	for flag, target := range map[string]*int{"--page": &q.Page, "--limit": &q.Limit} {
		if args.HasValue(flag) {
			n, err := strconv.ParseInt(args.Value(flag), 10, 32)
			if err != nil {
				return bad()
			}
			*target = int(n)
		}
	}
	if args.HasValue("--version") && q.Version == "" {
		return bad()
	}
	if (q.ResourceKind != "component" && q.ResourceKind != "backlog" && q.ResourceKind != "intake") || !q.ValidPage() {
		return bad()
	}
	return q, nil
}
