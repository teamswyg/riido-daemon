package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

// whoamiWorkspaces carries the configured default and looked-up names for the
// terminal view; both are optional.
type whoamiWorkspaces struct {
	defaultID string
	names     map[string]string
}

func whoamiView(w io.Writer, identity userapi.Identity, workspaces whoamiWorkspaces) textRenderer {
	return terminalView(w, func(out io.Writer, term cliui.Terminal) { renderWhoamiView(out, term, identity, workspaces) })
}

func workView(w io.Writer, work userapi.Work) textRenderer {
	return terminalView(w, func(out io.Writer, term cliui.Terminal) { renderWorkView(out, term, work) })
}

func relationsView(w io.Writer, relations []*userapi.WorkRelation) textRenderer {
	return terminalView(w, func(out io.Writer, term cliui.Terminal) { renderRelationsView(out, term, relations) })
}

func activityView(w io.Writer, page userapi.ActivityPage) textRenderer {
	return terminalView(w, func(out io.Writer, term cliui.Terminal) { renderActivityView(out, term, page) })
}

func renderWhoamiView(w io.Writer, term cliui.Terminal, id userapi.Identity, workspaces whoamiWorkspaces) {
	writeViewHeader(w, term, orText(orText(id.Username, id.Email), id.ID), "", term.Success("● 로그인됨"))
	fmt.Fprintln(w)
	fields := []viewField{
		{"이메일", id.Email},
		{"사용자 이름", id.Username},
		{"로그인 방식", id.LoginType},
		{"사용자 ID", id.ID},
		{"최근 워크스페이스", id.LastViewedWorkspaceID},
		{"기본 워크스페이스", workspaces.defaultID},
	}
	// Measure with the bare IDs, then shorten only the names so the full
	// workspace IDs stay copyable on narrow terminals.
	room := viewFieldRoom(term, viewFieldLabelWidth(fields))
	fields[4].value = workspaceLabel(id.LastViewedWorkspaceID, workspaces.names, room)
	fields[5].value = workspaceLabel(workspaces.defaultID, workspaces.names, room)
	writeViewFields(w, term, fields)
}

func renderWorkView(w io.Writer, term cliui.Terminal, work userapi.Work) {
	var status userapi.WorkStatus
	if work.Status != nil {
		status = *work.Status
	}
	meta := []string{viewState(term, status.StatusName, status.StatusType, work.IsDone)}
	if work.IsArchived {
		meta = append(meta, term.Warn("보관됨"))
	}
	if work.ComponentType != "" {
		meta = append(meta, term.Dim(oneLine(work.ComponentType)))
	}
	if work.UpdatedAt != "" {
		meta = append(meta, term.Dim("수정 "+viewTime(work.UpdatedAt)))
	}
	writeViewHeader(w, term, work.Title, work.KeyNumber, meta...)
	fmt.Fprintln(w)
	writeViewFields(w, term, []viewField{
		{"담당자", workMembers(work.ComponentMembers)},
		{"라벨", workLabels(work.ComponentLabels)},
		{"기간", viewPeriod(work.StartDate, work.EndDate)},
		{"추정", intText(work.Estimate)},
		{"우선순위 점수", intText(work.TotalPriorityScore)},
		{"만든 시각", dateText(work.CreatedAt)},
		{"ID", work.ID},
	})
	shown := writeViewSection(w, term, "설명", cliui.HTMLText(work.Description))
	if work.Document != nil && writeViewSection(w, term, "문서", cliui.HTMLText(work.Document.HTMLContent)) {
		shown = true
	}
	if !shown {
		fmt.Fprintln(w)
		fmt.Fprintln(w, term.Dim("  설명이 없어요."))
	}
}

func workMembers(members []userapi.WorkMember) string {
	names := make([]string, 0, len(members))
	for _, member := range members {
		if member.TeamMember != nil && member.TeamMember.WorkspaceProfile != nil && member.TeamMember.WorkspaceProfile.Nickname != "" {
			names = append(names, member.TeamMember.WorkspaceProfile.Nickname)
		}
	}
	return strings.Join(names, ", ")
}

func workLabels(labels []userapi.WorkLabel) string {
	names := make([]string, 0, len(labels))
	for _, label := range labels {
		if label.Label != nil && label.Label.LabelName != "" {
			names = append(names, label.Label.LabelName)
		}
	}
	return strings.Join(names, ", ")
}

func viewPeriod(start, end *string) string {
	if start == nil && end == nil {
		return ""
	}
	from, to := "", ""
	if start != nil {
		from = viewDate(*start, false)
	}
	if end != nil {
		to = viewDate(*end, false)
	}
	return strings.TrimSpace(from + " → " + to)
}

func intText(n *int) string {
	if n == nil {
		return ""
	}
	return strconv.Itoa(*n)
}

func dateText(ts string) string {
	if ts == "" {
		return ""
	}
	return viewDate(ts, true)
}

var relationLabels = map[userapi.RelationCategory]string{
	userapi.RelationRelatedTo:    "관련",
	userapi.RelationBlocking:     "막고 있음",
	userapi.RelationBlockedBy:    "막혀 있음",
	userapi.RelationHasDuplicate: "중복 있음",
	userapi.RelationDuplicatedOf: "중복 항목",
}

func renderRelationsView(w io.Writer, term cliui.Terminal, relations []*userapi.WorkRelation) {
	fmt.Fprintln(w, term.Bold(fmt.Sprintf("연결된 작업 %d개", len(relations))))
	if len(relations) == 0 {
		fmt.Fprintln(w, term.Dim("연결된 작업이 없어요."))
		return
	}
	width := 0
	for _, label := range relationLabels {
		width = max(width, cliui.DisplayWidth(label))
	}
	for _, relation := range relations {
		if relation == nil {
			continue
		}
		label := relationLabels[relation.RelationCategory]
		if label == "" {
			label = oneLine(string(relation.RelationCategory))
		}
		title, state := relatedIssueText(term, relation)
		room := max(term.Width-width-cliui.DisplayWidth(state)-4, 12)
		fmt.Fprintln(w, term.Accent(cliui.PadRight(label, width))+"  "+state+"  "+term.Bold(cliui.Truncate(orText(title, "(제목 없음)"), room)))
		fmt.Fprintln(w, strings.Repeat(" ", width+2)+term.Dim(oneLine(string(relation.IssueType))+" · "+oneLine(relation.IssueID)))
	}
}

func relatedIssueText(term cliui.Terminal, relation *userapi.WorkRelation) (string, string) {
	issue := relation.Issue
	if issue == nil {
		return "", term.Dim("○ 알 수 없음")
	}
	switch {
	case issue.Component != nil:
		c := issue.Component
		statusType := ""
		if c.Status != nil {
			statusType = deref(c.Status.StatusType)
		}
		return oneLine(deref(c.Title)), viewState(term, "", statusType, deref(c.IsDone))
	case issue.Backlog != nil:
		return oneLine(deref(issue.Backlog.Title)), term.Dim("○ 백로그")
	case issue.Intake != nil:
		return oneLine(deref(issue.Intake.Title)), term.Dim("○ 요청")
	}
	return "", term.Dim("○ 알 수 없음")
}

func renderActivityView(w io.Writer, term cliui.Terminal, page userapi.ActivityPage) {
	writeOffsetHeader(w, term, "활동", len(page.ActivityLogs), page.Meta)
	if len(page.ActivityLogs) == 0 {
		fmt.Fprintln(w, term.Dim("활동 기록이 없어요."))
		return
	}
	rows := make([][]string, 0, len(page.ActivityLogs))
	for _, entry := range page.ActivityLogs {
		if entry == nil {
			continue
		}
		rows = append(rows, []string{term.Dim(viewTime(deref(entry.CreatedAt))), oneLine(activityAuthor(entry.Author)), activityChange(entry)})
	}
	fmt.Fprintln(w)
	_ = term.Table(w, []string{"시각", "작성자", "변경"}, rows)
	writeOffsetNext(w, term, page.Meta)
}

func activityAuthor(author *userapi.ActivityAuthor) string {
	if author == nil || author.WorkspaceProfile == nil {
		return "-"
	}
	return orText(deref(author.WorkspaceProfile.Nickname), "-")
}

func activityChange(entry *userapi.ActivityLog) string {
	parts := make([]string, 0, 3)
	for _, value := range []*string{entry.Category, entry.Action, entry.Property} {
		if text := deref(value); text != "" {
			parts = append(parts, text)
		}
	}
	change := strings.Join(parts, " ")
	if old, next := deref(entry.OldValue), deref(entry.NewValue); old != "" || next != "" {
		change += ": " + orText(cliui.HTMLText(old), "-") + " → " + orText(cliui.HTMLText(next), "-")
	}
	return oneLine(change)
}

// writeOffsetHeader prints "<label> <total>건" with the page position.
func writeOffsetHeader(w io.Writer, term cliui.Terminal, label string, count int, meta *userapi.OffsetMeta) {
	total := count
	if meta != nil && meta.Total != nil {
		total = int(*meta.Total)
	}
	line := term.Bold(fmt.Sprintf("%s %d건", label, total))
	if meta != nil && meta.CurrentPage != nil && meta.LastPage != nil {
		line += " " + term.Dim(fmt.Sprintf("%d/%d 페이지", *meta.CurrentPage, max(*meta.LastPage, 1)))
	}
	fmt.Fprintln(w, line)
}

func writeOffsetNext(w io.Writer, term cliui.Terminal, meta *userapi.OffsetMeta) {
	if meta == nil {
		return
	}
	if next, ok := rawInt(meta.NextPage); ok {
		fmt.Fprintln(w)
		writeViewHint(w, term, fmt.Sprintf("다음 페이지: 같은 옵션에 --page %d", next))
	}
}
