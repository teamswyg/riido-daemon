package main

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (c contextIO) runUserMeetingList(raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := c.parseUserWorkArgs(raw, contextcli.ArgSpec{Values: []string{"--work", "--team-id", "--page", "--budget"}}, true)
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printMeetingUsage(c.stderr)
		return nil
	}
	if !args.HasValue("--work") {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "--work와 --team-id가 필요합니다", "riido meeting list --work <ID> --team-id <TEAM-ID>"))
	}
	if slices.ContainsFunc([]string{"--cursor", "--page", "--budget"}, args.HasValue) {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "미팅 목록은 범위·페이지 옵션을 지원하지 않습니다", "작업 ID와 팀 ID 또는 표시 키와 워크스페이스/팀 범위를 지정하세요. 직접 연결된 미팅을 서버가 최대 5개 선택합니다"))
	}
	work, team, cliErr := userWorkTarget(args, args.Value("--work"))
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	list, err := session.Meetings(ctx, work, team, args.Workspace)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	env := contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "live": true, "meetingList": list})
	if list.ContextStatus == "partial" || slices.ContainsFunc(list.Meetings, func(row userapi.MeetingListRow) bool { return row.Status == "partial" }) {
		env.Warnings = append(env.Warnings, contextcli.Warning{Code: "CONTEXT_INCOMPLETE", Message: "선택된 미팅 요약입니다. 원래 맥락에는 이어 읽을 내용 또는 조회하지 못한 자료가 있습니다"})
	}
	return c.succeed(args.JSON, env, func(w io.Writer, _ any) { renderUserMeetingList(w, list) })
}

func renderUserMeetingList(w io.Writer, list *userapi.MeetingList) {
	fmt.Fprintf(w, "작업 %q의 선택된 연관 미팅: %d개 (최대 %d개, 한 단계)\n", list.WorkID, len(list.Meetings), list.SelectionLimit)
	for _, meeting := range list.Meetings {
		fmt.Fprintf(w, "  - %q (id=%q, team=%q)\n", meeting.Title, meeting.ID, meeting.TeamID)
	}
	fmt.Fprintln(w, "현재 권한과 서버 선택 범위 안의 미팅 요약입니다. 전체 미팅 목록이나 본문을 조회한 결과는 아닙니다.")
}
