package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func bodyPageView(w io.Writer, page userapi.ReadBodyPage) textRenderer {
	return terminalView(w, func(out io.Writer, term cliui.Terminal) { renderBodyPageView(out, term, page) })
}

func contextView(w io.Writer, scope userapi.ReadScope) textRenderer {
	return terminalView(w, func(out io.Writer, term cliui.Terminal) { renderContextView(out, term, scope) })
}

func meetingView(w io.Writer, meeting userapi.ReadMeeting) textRenderer {
	return terminalView(w, func(out io.Writer, term cliui.Terminal) { renderMeetingView(out, term, meeting) })
}

var bodyPartLabels = map[string]string{
	"document": "본문", "description": "설명", "alternate": "대체 본문", "metadata": "메타데이터",
	"comments_page": "댓글 페이지", "board_page": "보드 페이지",
}

// renderBodyPageView shows one byte window of a body exactly as returned:
// the content can stop mid-markup, so it is printed raw (escaped and
// wrapped) rather than reformatted.
func renderBodyPageView(w io.Writer, term cliui.Terminal, page userapi.ReadBodyPage) {
	title := orText(bodyPartLabels[page.Part], page.Part)
	meta := []string{readStatusBadge(term, page.Status)}
	if page.Format != "" {
		meta = append(meta, term.Dim("형식 "+oneLine(page.Format)))
	}
	if page.Offset != nil {
		meta = append(meta, term.Dim(fmt.Sprintf("%d–%d 바이트", *page.Offset, *page.Offset+len(page.Content))))
	}
	writeViewHeader(w, term, title, page.ResourceID, meta...)
	fmt.Fprintln(w, term.Dim("version "+oneLine(page.Version)))
	fmt.Fprintln(w)
	if page.Content == "" {
		fmt.Fprintln(w, term.Dim("  (내용 없음)"))
	} else {
		writeViewBody(w, term, page.Content, 2)
	}
	fmt.Fprintln(w)
	switch {
	case page.NextOffset != nil:
		writeViewHint(w, term, fmt.Sprintf("다음 부분: 같은 옵션에 --version %s --offset %d", oneLine(page.Version), *page.NextOffset))
	case deref(page.Complete):
		writeViewHint(w, term, "끝까지 읽었어요.")
	}
}

// writeBodyPageHint points at a part that was too large to return inline.
// Its content is not shown, so the read restarts at the page's own offset and
// repeats its selector; otherwise the body read would resolve another page.
func writeBodyPageHint(w io.Writer, term cliui.Terminal, command string, page *userapi.ReadBodyPage) {
	hint := fmt.Sprintf("이어 읽기: %s --part %s --version %s", command, oneLine(page.Part), oneLine(page.Version))
	if offset := deref(page.Offset); offset > 0 {
		hint += fmt.Sprintf(" --offset %d", offset)
	}
	writeViewHint(w, term, hint+bodySelectorFlags(page.Part, page.Selector))
}

func bodySelectorFlags(part string, selector *userapi.ReadPartSelector) string {
	if selector == nil {
		return ""
	}
	flags := ""
	if part == "board_page" {
		if selector.BoardCursor != nil {
			flags += " --board-cursor " + oneLine(*selector.BoardCursor)
		}
		if selector.Limit != nil {
			flags += fmt.Sprintf(" --board-limit %d", *selector.Limit)
		}
		return flags
	}
	if selector.ParentCommentID != nil {
		flags += " --comment-id " + oneLine(*selector.ParentCommentID)
	}
	if selector.Page != nil {
		flags += fmt.Sprintf(" --page %d", *selector.Page)
	}
	if selector.Limit != nil {
		flags += fmt.Sprintf(" --comments-limit %d", *selector.Limit)
	}
	return flags
}

// scopeHeadline reads only the display fields every detail kind selects.
// status is an object for components and a string for intakes.
type scopeHeadline struct {
	Title     *string         `json:"title"`
	KeyNumber *string         `json:"keyNumber"`
	IsDone    *bool           `json:"isDone"`
	Status    json.RawMessage `json:"status"`
}

func workHeadline(term cliui.Terminal, work *userapi.ReadScopeWork) (string, string, string) {
	var head scopeHeadline
	for _, raw := range []json.RawMessage{work.Details, work.BacklogDetails, work.IntakeDetails} {
		if len(raw) > 0 && string(raw) != "null" && json.Unmarshal(raw, &head) == nil {
			break
		}
	}
	title, key := deref(head.Title), deref(head.KeyNumber)
	if work.Summary != nil {
		title, key = orText(title, work.Summary.Title), orText(key, deref(work.Summary.KeyNumber))
	}
	var status struct {
		StatusName string `json:"statusName"`
		StatusType string `json:"statusType"`
	}
	if json.Unmarshal(head.Status, &status) != nil {
		var name string
		if json.Unmarshal(head.Status, &name) == nil {
			status.StatusName = name
		}
	}
	state := ""
	if status.StatusName != "" || status.StatusType != "" || deref(head.IsDone) {
		state = viewState(term, status.StatusName, status.StatusType, deref(head.IsDone))
	}
	return title, key, state
}

func reasonText(reasons []*userapi.ReadScopeReason) string {
	parts := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		if reason == nil {
			continue
		}
		text := reason.Kind
		if relation := deref(reason.Relation); relation != "" {
			text += "(" + relation + ")"
		}
		parts = append(parts, text)
	}
	return oneLine(strings.Join(parts, ", "))
}

func renderContextView(w io.Writer, term cliui.Terminal, scope userapi.ReadScope) {
	counts := term.Dim(fmt.Sprintf("작업 %d개 · 미팅 %d개", len(scope.Works), len(scope.Meetings)))
	writeViewHeader(w, term, "작업 맥락", scope.RootID, readStatusBadge(term, scope.Status), counts)
	for _, work := range scope.Works {
		if work == nil {
			continue
		}
		title, key, state := workHeadline(term, work)
		fmt.Fprintln(w)
		line := term.Bold(cliui.Truncate(orText(oneLine(title), "(제목 없음)"), max(term.Width-cliui.DisplayWidth(key)-1, 12)))
		if key != "" {
			line += " " + term.Dim(oneLine(key))
		}
		fmt.Fprintln(w, line)
		meta := []string{state, term.Dim(oneLine(work.Kind)), term.Accent(reasonText(work.Reasons))}
		if work.ContentPage != nil {
			meta = append(meta, term.Warn("나눠 읽기"))
		} else if incompleteParts(work.PartsStatus) > 0 {
			meta = append(meta, term.Warn(fmt.Sprintf("일부 %d개 미완료", incompleteParts(work.PartsStatus))))
		}
		fmt.Fprintln(w, "  "+joinMeta(term, meta))
		fmt.Fprintln(w, "  "+term.Dim("id "+oneLine(work.ID)))
		if work.ContentPage != nil {
			writeBodyPageHint(w, term, workBodyCommand(work.ID, work.TeamID, work.Kind), work.ContentPage)
		}
	}
	if len(scope.Meetings) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, term.Dim("미팅"))
	}
	for _, meeting := range scope.Meetings {
		if meeting == nil {
			continue
		}
		title := ""
		if meeting.Snapshot != nil {
			title = meetingTitle(meeting.Snapshot)
		}
		fmt.Fprintln(w, "  "+joinMeta(term, []string{oneLine(title), term.Accent(reasonText(meeting.Reasons)), term.Dim("id " + oneLine(meeting.ID))}))
	}
	fmt.Fprintln(w)
	writeViewHint(w, term, "본문·댓글·첨부 등 전체 내용은 --json으로 볼 수 있어요.")
}

func incompleteParts(parts []*userapi.ReadPartStatus) int {
	count := 0
	for _, part := range parts {
		if part != nil && !deref(part.Complete) {
			count++
		}
	}
	return count
}

func joinMeta(term cliui.Terminal, parts []string) string {
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		if cliui.StripANSI(part) != "" {
			kept = append(kept, part)
		}
	}
	return strings.Join(kept, term.Dim(" · "))
}
