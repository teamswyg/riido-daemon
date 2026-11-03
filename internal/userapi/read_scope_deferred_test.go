package userapi

import (
	"encoding/json"
	"strings"
	"testing"
)

func deferredPart(id, part string, selector *ReadPartSelector) *ReadBodyPage {
	offset, next, complete := 0, 1, false
	return &ReadBodyPage{ResourceID: id, Part: part, Source: "riido_" + part, Format: "json", Version: "sha256:" + strings.Repeat("b", 64), Offset: &offset, Content: "{", NextOffset: &next, Complete: &complete, Status: "partial", Selector: selector}
}

func TestDeferredWorkPreservesReferenceInsteadOfClaimingEmptyMetadata(t *testing.T) {
	scope := readScopeFixture(t)
	w := scope.Works[0]
	w.Details = json.RawMessage("null")
	w.Summary = &ReadResourceSummary{ID: w.ID, Kind: w.Kind, Title: "큰 PR이 연결된 작업"}
	w.ContentPage = deferredPart(w.ID, "metadata", &ReadPartSelector{})
	if !validReadScope(scope, "root", "team") {
		t.Fatal("valid deferred metadata cannot be exposed for continuation")
	}
	w.Summary.ID = "foreign"
	if validReadScope(scope, "root", "team") {
		t.Fatal("deferred metadata accepted another resource summary")
	}
	w.Summary.ID = w.ID
	w.ContentPage.ResourceID = "foreign"
	if validReadScope(scope, "root", "team") {
		t.Fatal("deferred metadata accepted another resource body")
	}
}

func TestDeferredCommentsBindLogicalPageAndParent(t *testing.T) {
	for _, kind := range []string{"component", "meeting"} {
		t.Run(kind, func(t *testing.T) {
			page, limit, done, parent := 2, 3, false, "comment-one"
			q := ReadScopeCommentsRequest{ResourceID: "root", ResourceKind: kind, TeamID: "team", Page: page, Limit: limit, ParentCommentID: parent, Version: "sha256:" + strings.Repeat("a", 64)}
			p := &ReadScopeCommentPage{ResourceID: q.ResourceID, ResourceKind: q.ResourceKind, Version: q.Version, Status: "deferred", Page: page, Limit: limit, ParentCommentID: &parent, Complete: &done, Items: []json.RawMessage{}, ContentPage: deferredPart("root", "comments_page", &ReadPartSelector{Page: &page, Limit: &limit, ParentCommentID: &parent})}
			if !validReadScopeCommentPage(p, q) {
				t.Fatal("large logical comment page was rejected instead of returning its body reference")
			}
			other := "another-parent"
			p.ContentPage.Selector.ParentCommentID = &other
			if validReadScopeCommentPage(p, q) {
				t.Fatal("parent-bound comment continuation was substituted")
			}
		})
	}
}

func TestDeferredBoardKeepsWholeCardPageReachable(t *testing.T) {
	limit, offset, done := 100, 0, false
	emptyBody := readScopeFixture(t).Works[0].Body
	b := *emptyBody
	b.ResourceID = "meeting"
	d := b
	d.Part = "description"
	m := &ReadMeeting{ID: "meeting", TeamID: "team", Status: "partial", Details: json.RawMessage(`{"id":"meeting","meetingParticipants":[],"meetingAttendees":[]}`), Labels: []json.RawMessage{}, Body: &b, DescriptionBody: &d, PartsStatus: []*ReadPartStatus{}, Board: &ReadMeetingBoard{Version: strings.Repeat("a", 64), Offset: &offset, Complete: &done, Groups: []json.RawMessage{}, Cards: []json.RawMessage{}, ContentPage: deferredPart("meeting", "board_page", &ReadPartSelector{Limit: &limit})}}
	q := MeetingReadRequest{MeetingID: "meeting", TeamID: "team", BoardLimit: 100}
	if !validReadMeeting(m, q) {
		t.Fatal("large board page lost its typed continuation")
	}
	badCursor := "other-page"
	m.Board.ContentPage.Selector.BoardCursor = &badCursor
	if validReadMeeting(m, q) {
		t.Fatal("board reference accepted another logical page")
	}
}
