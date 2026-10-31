package main

import (
	"strconv"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (c contextIO) runUserMeetingComments(raw []string) error {
	id := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{MaxPositionals: 1, Values: []string{"--team-id", "--comment-id", "--page", "--version"}, AllowLimit: true})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), id, cliErr)
	}
	if args.Help {
		printMeetingUsage(c.stderr)
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, id, contextcli.Fail(contextcli.CodeUnsupported, "미팅 댓글은 사용자 모드에서 조회할 수 있습니다", "실행 자격을 사용자 자격으로 대체할 수 없습니다"))
	}
	q, cliErr := meetingCommentArguments(args)
	if cliErr != nil {
		return c.fail(args.JSON, id, cliErr)
	}
	return c.readUserScopeComments(id, args, q)
}

func meetingCommentArguments(args contextcli.Args) (userapi.ReadScopeCommentsRequest, *contextcli.CLIError) {
	q := userapi.ReadScopeCommentsRequest{ResourceKind: "meeting", Page: 1, Limit: 20, Version: args.Value("--version")}
	bad := func() (userapi.ReadScopeCommentsRequest, *contextcli.CLIError) {
		return userapi.ReadScopeCommentsRequest{}, contextcli.Fail(contextcli.CodeUsage, "미팅 ID·팀과 올바른 댓글 페이지가 필요합니다. 다음 페이지에는 원 version이 필요합니다", "riido meeting comments --help에서 --comment-id, --page, --version을 확인하세요")
	}
	if len(args.Positionals) != 1 || args.HasValue("--workspace") || args.HasValue("--cursor") {
		return bad()
	}
	var cliErr *contextcli.CLIError
	if q.ResourceID, cliErr = userCommentID(args.Positionals[0]); cliErr != nil {
		return q, cliErr
	}
	if q.TeamID, cliErr = userCommentID(args.Value("--team-id")); cliErr != nil {
		return q, cliErr
	}
	if args.HasValue("--comment-id") {
		if q.ParentCommentID, cliErr = userCommentID(args.Value("--comment-id")); cliErr != nil {
			return q, cliErr
		}
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
	if (args.HasValue("--version") && q.Version == "") || !q.Valid() {
		return bad()
	}
	return q, nil
}
