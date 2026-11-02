package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func commentsView(w io.Writer, page userapi.CommentPage) textRenderer {
	return terminalView(w, func(out io.Writer, term cliui.Terminal) { renderCommentsView(out, term, page) })
}

func repliesView(w io.Writer, replies []*userapi.Comment) textRenderer {
	return terminalView(w, func(out io.Writer, term cliui.Terminal) { renderRepliesView(out, term, replies) })
}

func renderCommentsView(w io.Writer, term cliui.Terminal, page userapi.CommentPage) {
	writeOffsetHeader(w, term, "댓글", len(page.Comments), page.Meta)
	writeCommentList(w, term, page.Comments, "댓글이 없어요.")
	writeOffsetNext(w, term, page.Meta)
}

func renderRepliesView(w io.Writer, term cliui.Terminal, replies []*userapi.Comment) {
	fmt.Fprintln(w, term.Bold(fmt.Sprintf("답글 %d건", len(replies))))
	writeCommentList(w, term, replies, "답글이 없어요.")
}

func scopeCommentsView(w io.Writer, page userapi.ReadScopeCommentPage) textRenderer {
	return terminalView(w, func(out io.Writer, term cliui.Terminal) { renderScopeCommentsView(out, term, page) })
}

func writeCommentList(w io.Writer, term cliui.Terminal, comments []*userapi.Comment, empty string) {
	shown := 0
	for _, comment := range comments {
		if comment == nil {
			continue
		}
		fmt.Fprintln(w)
		writeCommentView(w, term, comment)
		shown++
	}
	if shown == 0 {
		fmt.Fprintln(w, term.Dim(empty))
	}
}

// writeCommentView prints one comment like a conversation entry: author,
// time and badges on top, the HTML body as indented text below.
func writeCommentView(w io.Writer, term cliui.Terminal, comment *userapi.Comment) {
	author := "알 수 없음"
	if comment.Author != nil && comment.Author.WorkspaceProfile.Nickname != "" {
		author = comment.Author.WorkspaceProfile.Nickname
	}
	parts := []string{term.Bold(oneLine(author)), term.Dim(viewTime(comment.CreatedAt))}
	if comment.ReplyCount != nil && *comment.ReplyCount > 0 {
		parts = append(parts, term.Dim(fmt.Sprintf("답글 %d개", *comment.ReplyCount)))
	}
	if comment.IsFromSlack {
		parts = append(parts, term.Dim("Slack"))
	}
	if comment.IsFromDiscord {
		parts = append(parts, term.Dim("Discord"))
	}
	if comment.Status != "" && comment.Status != "active" {
		parts = append(parts, term.Warn(oneLine(comment.Status)))
	}
	parts = append(parts, term.Dim("id "+oneLine(comment.ID)))
	fmt.Fprintln(w, strings.Join(parts, term.Dim(" · ")))
	if body := cliui.HTMLText(comment.Body); strings.TrimSpace(body) != "" {
		writeViewBody(w, term, body, 2)
	} else {
		fmt.Fprintln(w, term.Dim("  (내용 없음)"))
	}
	for _, link := range []*string{comment.SlackMessageURL, comment.DiscordMessageURL} {
		if url := deref(link); url != "" {
			fmt.Fprintln(w, "  "+term.Dim(cliui.Truncate(oneLine(url), max(term.Width-2, 12))))
		}
	}
}

func renderScopeCommentsView(w io.Writer, term cliui.Terminal, page userapi.ReadScopeCommentPage) {
	label := "댓글"
	if page.ParentCommentID != nil {
		label = "답글"
	}
	fmt.Fprintln(w, term.Bold(fmt.Sprintf("%s %d건", label, len(page.Items)))+" "+term.Dim(fmt.Sprintf("%d페이지", page.Page)))
	fmt.Fprintln(w, readStatusBadge(term, page.Status)+term.Dim(" · version "+oneLine(page.Version)))
	if page.ContentPage != nil {
		fmt.Fprintln(w)
		writeViewHint(w, term, "댓글이 커서 나눠 읽어야 해요.")
		writeBodyPageHint(w, term, page.ContentPage)
		return
	}
	comments := make([]*userapi.Comment, 0, len(page.Items))
	unreadable := 0
	for _, raw := range page.Items {
		var comment userapi.Comment
		if err := json.Unmarshal(raw, &comment); err != nil {
			unreadable++
			continue
		}
		comments = append(comments, &comment)
	}
	writeCommentList(w, term, comments, label+"이 없어요.")
	if unreadable > 0 {
		writeViewHint(w, term, fmt.Sprintf("표시하지 못한 항목 %d개는 --json으로 확인하세요.", unreadable))
	}
	if page.NextPage != nil {
		fmt.Fprintln(w)
		writeViewHint(w, term, fmt.Sprintf("다음 페이지: 같은 옵션에 --page %d --version %s", *page.NextPage, oneLine(page.Version)))
	}
}
