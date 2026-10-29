package main

import (
	"context"
	"strconv"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (c contextIO) runUserBody(raw []string) error {
	return c.runUserResourceBody(raw, "component")
}

func (c contextIO) runUserResourceBody(raw []string, kind string) error {
	requestID := contextcli.NewRequestID()
	values := []string{"--team-id", "--part", "--version", "--offset", "--comment-id", "--page", "--comments-limit", "--board-cursor", "--board-limit"}
	if kind == "component" {
		values = append(values, "--kind")
	}
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{MaxPositionals: 1, Values: values, AllowLimit: true})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		if kind == "meeting" {
			printMeetingUsage(c.stderr)
		} else {
			printWorkUsage(c.stderr)
		}
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "에이전트는 저장된 내용 페이지를 조회해야 합니다", "riido context part <SOURCE-ID> --cursor <CURSOR>"))
	}
	q, cliErr := userResourceBodyArguments(args, kind)
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	page, err := session.Body(ctx, q)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "live": true, "body": page}), nil)
}

func userBodyArguments(args contextcli.Args) (userapi.ReadBodyRequest, *contextcli.CLIError) {
	return userResourceBodyArguments(args, "component")
}

func userResourceBodyArguments(args contextcli.Args, kind string) (userapi.ReadBodyRequest, *contextcli.CLIError) {
	q := userapi.ReadBodyRequest{Part: "document", Version: args.Value("--version"), Limit: userapi.DefaultReadBodyLimit}
	if kind == "meeting" {
		q.Kind = kind
	}
	if args.HasValue("--kind") {
		q.Kind = args.Value("--kind")
		if q.Kind != "component" && q.Kind != "backlog" && q.Kind != "intake" {
			return userapi.ReadBodyRequest{}, contextcli.Fail(contextcli.CodeUsage, "work body의 kind는 component, backlog, intake만 지원합니다", "미팅은 riido meeting body를 사용하세요")
		}
	}
	bad := func() (userapi.ReadBodyRequest, *contextcli.CLIError) {
		return userapi.ReadBodyRequest{}, contextcli.Fail(contextcli.CodeUsage, "자료·팀 ID, 원 본문 version과 올바른 페이지 범위가 필요합니다", "riido work body --help 또는 riido meeting body --help에서 자료 종류와 페이지 옵션을 확인하세요")
	}
	if len(args.Positionals) != 1 {
		return bad()
	}
	if args.HasValue("--cursor") || args.HasValue("--workspace") {
		return userapi.ReadBodyRequest{}, contextcli.Fail(contextcli.CodeUnsupported, "사용자 본문 조회는 --offset과 --limit을 사용합니다", "Agent의 snapshot cursor는 사용자 조회에 사용할 수 없습니다")
	}
	var err *contextcli.CLIError
	q.ComponentID, err = userCommentID(args.Positionals[0])
	if err != nil {
		return userapi.ReadBodyRequest{}, err
	}
	q.TeamID, err = userCommentID(args.Value("--team-id"))
	if err != nil {
		return userapi.ReadBodyRequest{}, err
	}
	if args.HasValue("--part") {
		q.Part = args.Value("--part")
	}
	for flag, target := range map[string]*int{"--offset": &q.Offset, "--limit": &q.Limit} {
		if args.HasValue(flag) {
			value, parseErr := strconv.ParseInt(args.Value(flag), 10, 32)
			if parseErr != nil {
				return bad()
			}
			*target = int(value)
		}
	}
	q.Selector, err = userBodySelector(args, q.Part)
	if err != nil {
		return userapi.ReadBodyRequest{}, err
	}
	if !q.Valid() {
		return bad()
	}
	return q, nil
}

func userBodySelector(args contextcli.Args, part string) (*userapi.ReadPartSelector, *contextcli.CLIError) {
	bad := func() (*userapi.ReadPartSelector, *contextcli.CLIError) {
		return nil, contextcli.Fail(contextcli.CodeUsage, "자료 종류에 맞는 원 페이지 선택값이 필요합니다", "comments_page에는 --comment-id·--page·--comments-limit, board_page에는 --board-cursor·--board-limit을 사용하세요")
	}
	s := &userapi.ReadPartSelector{}
	for _, flag := range []string{"--comment-id", "--page", "--comments-limit", "--board-cursor", "--board-limit"} {
		if !args.HasValue(flag) {
			continue
		}
		comments := flag == "--comment-id" || flag == "--page" || flag == "--comments-limit"
		if (comments && part != "comments_page") || (!comments && part != "board_page") {
			return bad()
		}
		value := args.Value(flag)
		switch flag {
		case "--comment-id":
			id, err := userCommentID(value)
			if err != nil {
				return nil, err
			}
			s.ParentCommentID = &id
		case "--board-cursor":
			if value == "" {
				return bad()
			}
			s.BoardCursor = &value
		default:
			n, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return bad()
			}
			v := int(n)
			if flag == "--page" {
				s.Page = &v
			} else {
				s.Limit = &v
			}
		}
	}
	if part == "metadata" || part == "comments_page" || part == "board_page" {
		return s, nil
	}
	return nil, nil
}
