package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

type viewMember struct {
	TeamMember *struct {
		WorkspaceProfile *struct {
			Nickname string `json:"nickname"`
		} `json:"workspaceProfile"`
	} `json:"teamMember"`
	WorkspaceProfile *struct {
		Nickname string `json:"nickname"`
	} `json:"workspaceProfile"`
}

func (m viewMember) name() string {
	switch {
	case m.TeamMember != nil && m.TeamMember.WorkspaceProfile != nil:
		return m.TeamMember.WorkspaceProfile.Nickname
	case m.WorkspaceProfile != nil:
		return m.WorkspaceProfile.Nickname
	}
	return ""
}

// meetingHeadline reads the display fields selected by the meeting query.
type meetingHeadline struct {
	Title           *string      `json:"title"`
	StartTime       *string      `json:"startTime"`
	EndTime         *string      `json:"endTime"`
	IsAllDay        *bool        `json:"isAllDay"`
	MeetingLocation *string      `json:"meetingLocation"`
	Attendees       []viewMember `json:"meetingAttendees"`
	CreatedBy       *viewMember  `json:"createdBy"`
}

type meetingCard struct {
	GroupID    string      `json:"groupId"`
	Body       *string     `json:"body"`
	TeamMember *viewMember `json:"teamMember"`
	Component  *struct {
		KeyNumber *string `json:"keyNumber"`
		Title     *string `json:"title"`
	} `json:"component"`
}

func renderMeetingView(w io.Writer, term cliui.Terminal, meeting userapi.ReadMeeting) {
	var head meetingHeadline
	if len(meeting.Details) > 0 {
		_ = json.Unmarshal(meeting.Details, &head)
	}
	title := deref(head.Title)
	if meeting.Summary != nil {
		title = orText(title, meeting.Summary.Title)
	}
	writeViewHeader(w, term, title, "", readStatusBadge(term, meeting.Status))
	fmt.Fprintln(w)
	creator := ""
	if head.CreatedBy != nil {
		creator = head.CreatedBy.name()
	}
	writeViewFields(w, term, []viewField{
		{"일시", meetingTime(head)},
		{"장소", deref(head.MeetingLocation)},
		{"참석자", memberNames(head.Attendees)},
		{"라벨", meetingLabels(meeting.Labels)},
		{"만든 사람", creator},
		{"ID", meeting.ID},
		{"팀 ID", meeting.TeamID},
	})
	if meeting.ContentPage != nil {
		fmt.Fprintln(w)
		writeViewHint(w, term, "미팅 내용이 커서 나눠 읽어야 해요.")
		writeBodyPageHint(w, term, "riido meeting body "+oneLine(meeting.ID)+" --team-id "+oneLine(meeting.TeamID), meeting.ContentPage)
		return
	}
	writeMeetingBody(w, term, meeting)
	writeMeetingBoard(w, term, meeting.Board)
	fmt.Fprintln(w)
	writeViewHint(w, term, "전체 내용은 --json으로 볼 수 있어요.")
}

func meetingTime(head meetingHeadline) string {
	if head.StartTime == nil {
		return ""
	}
	if deref(head.IsAllDay) {
		return viewDate(*head.StartTime, false) + " (종일)"
	}
	text := viewDate(*head.StartTime, true)
	if head.EndTime != nil {
		text += " → " + viewDate(*head.EndTime, true)
	}
	return text
}

func memberNames(members []viewMember) string {
	names := make([]string, 0, len(members))
	for _, member := range members {
		if name := member.name(); name != "" {
			names = append(names, name)
		}
	}
	return strings.Join(names, ", ")
}

func meetingLabels(labels []json.RawMessage) string {
	names := make([]string, 0, len(labels))
	for _, raw := range labels {
		var label struct {
			LabelName string `json:"labelName"`
		}
		if json.Unmarshal(raw, &label) == nil && label.LabelName != "" {
			names = append(names, label.LabelName)
		}
	}
	return strings.Join(names, ", ")
}

// writeMeetingBody shows the first body page; HTML is reflowed as text.
func writeMeetingBody(w io.Writer, term cliui.Terminal, meeting userapi.ReadMeeting) {
	body := meeting.Body
	if body == nil || body.Content == "" {
		return
	}
	switch body.Format {
	case "html":
		writeViewSection(w, term, "본문", cliui.HTMLText(body.Content))
	case "json":
		fmt.Fprintln(w)
		writeViewHint(w, term, fmt.Sprintf("본문은 %s 형식이라 riido meeting body 또는 --json으로 확인하세요.", oneLine(body.Format)))
	default:
		writeViewSection(w, term, "본문", body.Content)
	}
	if body.NextOffset != nil {
		writeViewHint(w, term, fmt.Sprintf("  이어 읽기: riido meeting body %s --team-id %s --version %s --offset %d",
			oneLine(meeting.ID), oneLine(meeting.TeamID), oneLine(body.Version), *body.NextOffset))
	}
}

func writeMeetingBoard(w io.Writer, term cliui.Terminal, board *userapi.ReadMeetingBoard) {
	if board == nil {
		return
	}
	groups := meetingGroups(board.Groups)
	fmt.Fprintln(w)
	fmt.Fprintln(w, term.Dim(fmt.Sprintf("보드 카드 %d개", len(board.Cards))))
	for _, raw := range board.Cards {
		var card meetingCard
		if json.Unmarshal(raw, &card) != nil {
			continue
		}
		fmt.Fprintln(w, "  "+meetingCardLine(term, card, groups))
	}
	if board.NextCursor != nil {
		writeViewHint(w, term, "  다음 보드 페이지: 같은 옵션에 --cursor "+oneLine(*board.NextCursor))
	}
}

func meetingCardLine(term cliui.Terminal, card meetingCard, groups map[string]string) string {
	parts := make([]string, 0, 4)
	if group := groups[card.GroupID]; group != "" {
		parts = append(parts, term.Dim("["+oneLine(group)+"]"))
	}
	if card.TeamMember != nil && card.TeamMember.name() != "" {
		parts = append(parts, term.Bold(oneLine(card.TeamMember.name())))
	}
	text := oneLine(cliui.HTMLText(deref(card.Body)))
	if card.Component != nil {
		text = strings.TrimSpace(text + " (" + oneLine(deref(card.Component.KeyNumber)+" "+deref(card.Component.Title)) + ")")
	}
	prefix := "• " + strings.Join(parts, " ")
	room := max(term.Width-cliui.DisplayWidth(prefix)-3, 12)
	return strings.TrimRight(prefix+" "+cliui.Truncate(text, room), " ")
}

// meetingGroups maps both the board-group link id and the group id to a name.
func meetingGroups(raws []json.RawMessage) map[string]string {
	names := map[string]string{}
	for _, raw := range raws {
		var group struct {
			ID    string `json:"id"`
			Group *struct {
				ID        string `json:"id"`
				GroupName string `json:"groupName"`
			} `json:"meetingBoardGroup"`
		}
		if json.Unmarshal(raw, &group) != nil || group.Group == nil {
			continue
		}
		names[group.ID] = group.Group.GroupName
		names[group.Group.ID] = group.Group.GroupName
	}
	return names
}
