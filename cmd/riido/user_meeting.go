package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (c contextIO) runUserMeeting(raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{MaxPositionals: 1, Values: []string{"--team-id"}, AllowLimit: true})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printMeetingUsage(c.stderr)
		return nil
	}
	q, cliErr := userMeetingArguments(args)
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	meeting, err := session.Meeting(ctx, q)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	env := contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "live": true, "meeting": meeting})
	if meeting.Status == "partial" {
		env.Warnings = append(env.Warnings, contextcli.Warning{Code: "CONTEXT_INCOMPLETE", Message: "미팅에 이어 읽을 내용 또는 조회하지 못한 자료가 있습니다. 본문·보드 페이지와 자료 상태를 확인하세요"})
	}
	if meetingHasDeferredContent(meeting) {
		env.Warnings = append(env.Warnings, deferredContentWarning())
	}
	return c.succeed(args.JSON, env, nil)
}

func userMeetingArguments(args contextcli.Args) (userapi.MeetingReadRequest, *contextcli.CLIError) {
	q := userapi.MeetingReadRequest{BoardCursor: args.Cursor, BoardLimit: 100}
	bad := func() (userapi.MeetingReadRequest, *contextcli.CLIError) {
		return userapi.MeetingReadRequest{}, contextcli.Fail(contextcli.CodeUsage, "미팅·팀 ID와 올바른 보드 페이지가 필요합니다", "riido meeting get <ID> --team-id <TEAM-ID> [--cursor CURSOR] [--limit 1..200]")
	}
	if len(args.Positionals) != 1 {
		return bad()
	}
	if args.HasValue("--workspace") {
		return userapi.MeetingReadRequest{}, contextcli.Fail(contextcli.CodeUnsupported, "미팅 조회에는 실제 팀 ID를 지정하세요", "--workspace 대신 --team-id를 사용하세요")
	}
	var err *contextcli.CLIError
	q.MeetingID, err = userCommentID(args.Positionals[0])
	if err != nil {
		return userapi.MeetingReadRequest{}, err
	}
	q.TeamID, err = userCommentID(args.Value("--team-id"))
	if err != nil {
		return userapi.MeetingReadRequest{}, err
	}
	if args.HasValue("--cursor") && strings.TrimSpace(q.BoardCursor) == "" {
		return bad()
	}
	if args.HasValue("--limit") {
		n, e := strconv.ParseInt(args.Value("--limit"), 10, 32)
		if e != nil {
			return bad()
		}
		q.BoardLimit = int(n)
	}
	if !q.Valid() {
		return bad()
	}
	return q, nil
}
