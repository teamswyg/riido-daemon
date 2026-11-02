package main

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

const viewEvil = "악성\x1b[2J\u202e제목\a"

var viewClock = time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)

func viewAgo(d time.Duration) string { return viewClock.Add(-d).Format(time.RFC3339Nano) }

func pinViewClock(t *testing.T) {
	t.Helper()
	previous := viewNow
	viewNow = func() time.Time { return viewClock }
	t.Cleanup(func() { viewNow = previous })
}

type viewCase struct {
	name   string
	render func(io.Writer, cliui.Terminal)
	want   []string
}

// checkViews renders every case with and without color and checks content,
// escaping of the injected controls and the terminal width.
func checkViews(t *testing.T, cases []viewCase) {
	t.Helper()
	pinViewClock(t)
	for _, tc := range cases {
		for _, color := range []bool{false, true} {
			term := cliui.Terminal{Interactive: true, Color: color, Width: 100}
			var buf bytes.Buffer
			tc.render(&buf, term)
			out := buf.String()
			if strings.Contains(out, "\x1b[2J") || strings.ContainsAny(out, "\a\u202e") {
				t.Fatalf("%s color=%t kept a raw control: %q", tc.name, color, out)
			}
			if !color && strings.Contains(out, "\x1b") {
				t.Fatalf("%s styled without color: %q", tc.name, out)
			}
			plain := cliui.StripANSI(out)
			for _, want := range tc.want {
				if !strings.Contains(plain, want) {
					t.Fatalf("%s color=%t missing %q in:\n%s", tc.name, color, want, plain)
				}
			}
			for line := range strings.SplitSeq(strings.TrimRight(out, "\n"), "\n") {
				// Command hints stay whole so they can be copied; the terminal wraps them.
				if cliui.DisplayWidth(line) > term.Width && !strings.Contains(line, " --") {
					t.Fatalf("%s line is %d cells: %q", tc.name, cliui.DisplayWidth(line), cliui.StripANSI(line))
				}
			}
		}
	}
}

func TestTerminalViewKeepsJSONOffTerminal(t *testing.T) {
	called := false
	if render := terminalView(&bytes.Buffer{}, func(io.Writer, cliui.Terminal) { called = true }); render != nil || called {
		t.Fatal("a buffer must keep the nil renderer and JSON output")
	}
	for _, render := range []textRenderer{
		whoamiView(io.Discard, userapi.Identity{}), workView(io.Discard, userapi.Work{}),
		relationsView(io.Discard, nil), activityView(io.Discard, userapi.ActivityPage{}),
		commentsView(io.Discard, userapi.CommentPage{}), repliesView(io.Discard, nil),
		scopeCommentsView(io.Discard, userapi.ReadScopeCommentPage{}), bodyPageView(io.Discard, userapi.ReadBodyPage{}),
		contextView(io.Discard, userapi.ReadScope{}), meetingView(io.Discard, userapi.ReadMeeting{}),
	} {
		if render != nil {
			t.Fatal("non-terminal writer got a terminal view")
		}
	}
}

func TestViewKeepsCopyableValuesWhole(t *testing.T) {
	term := cliui.Terminal{Interactive: true, Width: 30}
	id, cursor := strings.Repeat("a", 40), strings.Repeat("c", 60)
	var buf bytes.Buffer
	writeViewFields(&buf, term, []viewField{{"ID", id}, {"설명", strings.Repeat("긴 설명 ", 20)}})
	writeViewHint(&buf, term, "다음 페이지: 같은 옵션에 --cursor "+cursor)
	out := buf.String()
	if !strings.Contains(out, id) || !strings.Contains(out, cursor) {
		t.Fatalf("IDs and cursors must not be shortened:\n%s", out)
	}
	if !strings.Contains(out, "…") {
		t.Fatalf("prose was not shortened:\n%s", out)
	}
}

func TestBodyPageHintRestartsAtShownOffset(t *testing.T) {
	offset, next, limit, cursor := 4096, 8192, 30, "board-next"
	var buf bytes.Buffer
	term := cliui.Terminal{Interactive: true, Width: 200}
	writeBodyPageHint(&buf, term, "riido meeting body m --team-id t", &userapi.ReadBodyPage{Part: "metadata", Version: "v1", Offset: new(int), NextOffset: &next})
	writeBodyPageHint(&buf, term, "riido meeting body m --team-id t", &userapi.ReadBodyPage{
		Part: "board_page", Version: "v2", Offset: &offset, NextOffset: &next,
		Selector: &userapi.ReadPartSelector{BoardCursor: &cursor, Limit: &limit},
	})
	want := "이어 읽기: riido meeting body m --team-id t --part metadata --version v1\n" +
		"이어 읽기: riido meeting body m --team-id t --part board_page --version v2 --offset 4096 --board-cursor board-next --board-limit 30\n"
	if buf.String() != want {
		t.Fatalf("hint = %q, want %q", buf.String(), want)
	}
}

func TestMeetingContinuationsStartAtUnshownBytes(t *testing.T) {
	zero, next, limit := 0, 300, 100
	meeting := userapi.ReadMeeting{
		ID: "m2", TeamID: "t2", Status: "partial",
		Body: &userapi.ReadBodyPage{Part: "document", Format: "json", Version: "v6", Offset: &zero, NextOffset: &next, Content: `{"type":"doc"}`},
		Board: &userapi.ReadMeetingBoard{ContentPage: &userapi.ReadBodyPage{
			Part: "board_page", Version: "v7", Offset: &zero, Selector: &userapi.ReadPartSelector{Limit: &limit},
		}},
	}
	var buf bytes.Buffer
	renderMeetingView(&buf, cliui.Terminal{Interactive: true, Width: 200}, meeting)
	out := buf.String()
	for _, want := range []string{
		"이어 읽기: riido meeting body m2 --team-id t2 --part document --version v6\n",
		"보드 내용이 커서 나눠 읽어야 해요.",
		"이어 읽기: riido meeting body m2 --team-id t2 --part board_page --version v7 --board-limit 100\n",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "--offset") || strings.Contains(out, "보드 카드 0개") || strings.Contains(out, `"type"`) {
		t.Fatalf("unshown bytes were skipped or a deferred board looked empty:\n%s", out)
	}
}

func TestWhoamiAndWorkViews(t *testing.T) {
	start, end, estimate := "2026-10-01T03:00:00.000Z", "2026-10-31T03:00:00.000Z", 3
	member := userapi.WorkMember{TeamMember: &userapi.WorkTeamMember{WorkspaceProfile: &userapi.WorkWorkspaceProfile{Nickname: "리도"}}}
	work := userapi.Work{
		ID: "work-1", Title: viewEvil, KeyNumber: "RIID-12", ComponentType: "task", UpdatedAt: viewAgo(3 * time.Hour),
		Status: &userapi.WorkStatus{StatusName: "진행 중", StatusType: "inProgress"}, StartDate: &start, EndDate: &end, Estimate: &estimate,
		ComponentMembers: []userapi.WorkMember{member, {}}, ComponentLabels: []userapi.WorkLabel{{Label: &userapi.WorkLabelInfo{LabelName: "버그"}}},
		Description: "<p>설명 <b>굵게</b></p>",
		Document:    &userapi.Document{HTMLContent: "<p>" + strings.Repeat("아주 긴 문서 문장입니다 ", 20) + "</p>"},
	}
	checkViews(t, []viewCase{
		{"whoami", func(w io.Writer, term cliui.Terminal) {
			renderWhoamiView(w, term, userapi.Identity{ID: "user-1", Email: "me@example.com", Username: viewEvil, LoginType: "google"})
		}, []string{cliui.Safe(viewEvil), "로그인됨", "me@example.com", "user-1", "google"}},
		{
			"work", func(w io.Writer, term cliui.Terminal) { renderWorkView(w, term, work) },
			[]string{"RIID-12", "● 진행 중", "수정 3시간 전", "리도", "버그", "2026-10-01 → 2026-10-31", "설명 굵게", "아주 긴 문서", "work-1"},
		},
		{"done work without text", func(w io.Writer, term cliui.Terminal) {
			renderWorkView(w, term, userapi.Work{ID: "w", IsDone: true, IsArchived: true})
		}, []string{"(제목 없음)", "● 완료", "보관됨", "설명이 없어요."}},
	})
}

func TestRelationsAndActivityViews(t *testing.T) {
	done, title, backlog := true, viewEvil, "백로그 제목"
	relations := []*userapi.WorkRelation{
		{RelationCategory: userapi.RelationBlockedBy, IssueID: "c1", IssueType: userapi.RelatedIssueComponent, Issue: &userapi.RelatedIssue{Component: &userapi.RelatedComponent{Title: &title, IsDone: &done}}},
		{RelationCategory: userapi.RelationRelatedTo, IssueID: "b1", IssueType: userapi.RelatedIssueBacklog, Issue: &userapi.RelatedIssue{Backlog: &userapi.RelatedTitledResource{Title: &backlog}}},
		nil,
	}
	total, current, last := int32(3), int32(1), int32(2)
	action, property, old, next, nick := "update", "title", "<p>옛 제목</p>", viewEvil, "작성자"
	createdAt := viewAgo(5 * time.Minute)
	page := userapi.ActivityPage{
		ActivityLogs: []*userapi.ActivityLog{{
			ID: "a1", CreatedAt: &createdAt, Action: &action, Property: &property, OldValue: &old, NewValue: &next,
			Author: &userapi.ActivityAuthor{WorkspaceProfile: &userapi.ActivityProfile{Nickname: &nick}},
		}, nil},
		Meta: &userapi.OffsetMeta{Total: &total, CurrentPage: &current, LastPage: &last, NextPage: json.RawMessage("2"), PrevPage: json.RawMessage("null")},
	}
	checkViews(t, []viewCase{
		{
			"relations", func(w io.Writer, term cliui.Terminal) { renderRelationsView(w, term, relations) },
			[]string{"연결된 작업 3개", "막혀 있음", "● 완료", `악성\x1b[2J`, "component · c1", "관련", "백로그 제목", "backlog · b1"},
		},
		{"no relations", func(w io.Writer, term cliui.Terminal) { renderRelationsView(w, term, nil) }, []string{"연결된 작업이 없어요."}},
		{
			"activity", func(w io.Writer, term cliui.Terminal) { renderActivityView(w, term, page) },
			[]string{"활동 3건", "1/2 페이지", "시각", "5분 전", "작성자", "update title: 옛 제목 → 악성", "--page 2"},
		},
		{"no activity", func(w io.Writer, term cliui.Terminal) { renderActivityView(w, term, userapi.ActivityPage{}) }, []string{"활동 0건", "활동 기록이 없어요."}},
	})
}

func viewComment(t *testing.T) *userapi.Comment {
	t.Helper()
	replies, slack := int32(2), "https://slack.example/m\x1b[2J"
	comment := &userapi.Comment{
		ID: "comment-1", Status: "active", CreatedAt: viewAgo(26 * time.Hour), Body: "<p>안녕\a</p><ul><li>하나</li></ul>",
		ReplyCount: &replies, IsFromSlack: true, SlackMessageURL: &slack,
	}
	comment.Author = &userapi.CommentAuthor{}
	comment.Author.WorkspaceProfile.Nickname = viewEvil
	return comment
}

func TestCommentViews(t *testing.T) {
	comment := viewComment(t)
	raw, err := json.Marshal(comment)
	if err != nil {
		t.Fatal(err)
	}
	total, current, last, nextPage, parent := int32(1), int32(1), int32(1), 2, "comment-1"
	complete := false
	scope := userapi.ReadScopeCommentPage{
		ResourceID: "work-1", Version: "v1", Status: "partial", Page: 1, NextPage: &nextPage, Complete: &complete,
		Items: []json.RawMessage{raw, json.RawMessage(`"bad"`)},
	}
	replies := scope
	replies.ParentCommentID, replies.NextPage = &parent, nil
	deferredPage, deferredLimit, deferredNext, zero := 2, 50, 100, 0
	deferred := userapi.ReadScopeCommentPage{
		ResourceID: "work-1", ResourceKind: "backlog", Version: "v1", Status: "deferred", Page: 1,
		ContentPage: &userapi.ReadBodyPage{
			Part: "comments_page", Version: "v2", Offset: &zero, NextOffset: &deferredNext,
			Selector: &userapi.ReadPartSelector{ParentCommentID: &parent, Page: &deferredPage, Limit: &deferredLimit},
		},
	}
	checkViews(t, []viewCase{
		{"comments", func(w io.Writer, term cliui.Terminal) {
			meta := &userapi.OffsetMeta{Total: &total, CurrentPage: &current, LastPage: &last, NextPage: json.RawMessage("null")}
			renderCommentsView(w, term, userapi.CommentPage{Comments: []*userapi.Comment{comment, nil}, Meta: meta})
		}, []string{"댓글 1건", "1/1 페이지", `악성\x1b[2J`, "1일 전", "답글 2개", "Slack", "id comment-1", `안녕\a`, "• 하나", `https://slack.example/m\x1b[2J`}},
		{"no comments", func(w io.Writer, term cliui.Terminal) { renderCommentsView(w, term, userapi.CommentPage{}) }, []string{"댓글 0건", "댓글이 없어요."}},
		{"replies", func(w io.Writer, term cliui.Terminal) { renderRepliesView(w, term, []*userapi.Comment{comment}) }, []string{"답글 1건", `안녕\a`}},
		{"no replies", func(w io.Writer, term cliui.Terminal) { renderRepliesView(w, term, nil) }, []string{"답글이 없어요."}},
		{
			"scope comments", func(w io.Writer, term cliui.Terminal) { renderScopeCommentsView(w, term, scope) },
			[]string{"댓글 2건", "● 일부", "version v1", `안녕\a`, "표시하지 못한 항목 1개", "--page 2 --version v1"},
		},
		{"scope replies", func(w io.Writer, term cliui.Terminal) { renderScopeCommentsView(w, term, replies) }, []string{"답글 2건"}},
		{
			"deferred comments", func(w io.Writer, term cliui.Terminal) { renderScopeCommentsView(w, term, deferred) },
			[]string{"나눠 읽기", "나눠 읽어야", "riido work body work-1 --team-id <TEAM-ID> --kind backlog --part comments_page --version v2 --comment-id comment-1 --page 2 --comments-limit 50"},
		},
	})
}

func TestBodyContextAndMeetingViews(t *testing.T) {
	offset, next, complete := 0, 400, false
	body := userapi.ReadBodyPage{
		ResourceID: "work-1", Part: "document", Status: "partial", Format: "json", Version: "v3", Offset: &offset, NextOffset: &next, Complete: &complete,
		Content: "{\"type\":\"doc\",\"text\":\"" + strings.Repeat("가나다", 60) + "\"}\n" + viewEvil,
	}
	keyNumber, relation := "RIID-9", "blocking"
	scope := userapi.ReadScope{RootID: "work-1", Status: "partial", Works: []*userapi.ReadScopeWork{
		{ID: "work-1", Kind: "component", Reasons: []*userapi.ReadScopeReason{{Kind: "root"}}, Details: json.RawMessage(`{"title":"루트 작업","keyNumber":"RIID-1","isDone":true,"status":{"statusName":"완료","statusType":"done"}}`)},
		{
			ID: "intake-1", Kind: "intake", Reasons: []*userapi.ReadScopeReason{{Kind: "relation", Relation: &relation}}, Details: json.RawMessage("null"), IntakeDetails: json.RawMessage(`{"title":"요청 \u001b[2J","status":"open"}`),
			PartsStatus: []*userapi.ReadPartStatus{{Name: "comments", Complete: &complete}},
		},
		{ID: "work-9", TeamID: "team-1", Kind: "component", Summary: &userapi.ReadResourceSummary{Title: "요약 제목", KeyNumber: &keyNumber}, ContentPage: &userapi.ReadBodyPage{Part: "metadata", Version: "v8"}},
		nil,
	}, Meetings: []*userapi.ReadMeetingReference{{ID: "meeting-1", Reasons: []*userapi.ReadScopeReason{{Kind: "linked"}}, Snapshot: &userapi.ReadMeeting{Summary: &userapi.ReadResourceSummary{Title: "주간 회의"}}}}}
	cursor := "next-cursor"
	meeting := userapi.ReadMeeting{
		ID: "meeting-1", TeamID: "team-1", Status: "available",
		Details: json.RawMessage(`{"title":"주간 회의 \u001b[2J","startTime":"2026-11-02T01:00:00.000Z","endTime":"2026-11-02T02:00:00.000Z","isAllDay":false,"meetingLocation":"회의실",
			"meetingAttendees":[{"teamMember":{"workspaceProfile":{"nickname":"리도"}}}],"createdBy":{"workspaceProfile":{"nickname":"코덱스"}}}`),
		Labels: []json.RawMessage{json.RawMessage(`{"labelName":"정기"}`)},
		Body:   &userapi.ReadBodyPage{Format: "html", Part: "document", Version: "v5", Content: "<h1>안건</h1><p>첫 안건</p>", NextOffset: &next},
		Board: &userapi.ReadMeetingBoard{
			NextCursor: &cursor,
			Groups:     []json.RawMessage{json.RawMessage(`{"id":"link-1","meetingBoardGroup":{"id":"group-1","groupName":"할 일"}}`)},
			Cards:      []json.RawMessage{json.RawMessage(`{"groupId":"group-1","body":"<p>카드 내용</p>","teamMember":{"workspaceProfile":{"nickname":"리도"}},"component":{"keyNumber":"RIID-3","title":"연결 작업"}}`), json.RawMessage(`1`)},
		},
	}
	checkViews(t, []viewCase{
		{
			"body", func(w io.Writer, term cliui.Terminal) { renderBodyPageView(w, term, body) },
			[]string{"본문 work-1", "● 일부", "형식 json", "0–", "version v3", "가나다가나다", `악성\x1b[2J`, "--version v3 --offset 400"},
		},
		{"empty body", func(w io.Writer, term cliui.Terminal) {
			done := true
			renderBodyPageView(w, term, userapi.ReadBodyPage{Part: "description", Status: "empty", Complete: &done})
		}, []string{"설명", "비어 있음", "(내용 없음)", "끝까지 읽었어요."}},
		{
			"context", func(w io.Writer, term cliui.Terminal) { renderContextView(w, term, scope) },
			[]string{"작업 맥락 work-1", "작업 4개 · 미팅 1개", "루트 작업 RIID-1", "● 완료", "root", `요청 \x1b[2J`, "○ open", "relation(blocking)", "일부 1개 미완료", "요약 제목 RIID-9", "나눠 읽기", "이어 읽기: riido work body work-9 --team-id team-1 --part metadata --version v8", "주간 회의", "id meeting-1", "--json"},
		},
		{
			"meeting", func(w io.Writer, term cliui.Terminal) { renderMeetingView(w, term, meeting) },
			[]string{
				`주간 회의 \x1b[2J`, "● 전체", "→", "회의실", "참석자", "리도", "정기", "코덱스", "team-1", "안건", "첫 안건",
				"이어 읽기: riido meeting body meeting-1 --team-id team-1 --part document --version v5 --offset 400", "보드 카드 2개", "[할 일] 리도 카드 내용 (RIID-3 연결 작업)", "--cursor next-cursor",
			},
		},
		{"deferred meeting", func(w io.Writer, term cliui.Terminal) {
			renderMeetingView(w, term, userapi.ReadMeeting{ID: "m", Status: "partial", ContentPage: &userapi.ReadBodyPage{Part: "metadata", Version: "v9"}})
		}, []string{"(제목 없음)", "나눠 읽어야", "--part metadata --version v9"}},
	})
}
