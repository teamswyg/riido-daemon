package userapi

func validReadSummary(summary *ReadResourceSummary, id, kind string) bool {
	return summary != nil && summary.ID == id && summary.Kind == kind
}

func validDeferredPart(body *ReadBodyPage, id, kind, part string, selector *ReadPartSelector) bool {
	if body == nil || body.Offset == nil || *body.Offset != 0 {
		return false
	}
	q := ReadBodyRequest{ComponentID: id, TeamID: "scope-authorized", Kind: kind, Part: part, Version: body.Version, Limit: MaxReadBodyLimit, Selector: selector}
	return q.Valid() && validReadBodyPage(body, q)
}

func validDeferredComments(page *ReadScopeCommentPage) bool {
	if page == nil || page.ContentPage == nil || page.Complete == nil || *page.Complete || page.NextPage != nil || page.Status != "deferred" || page.Items == nil || len(page.Items) != 0 {
		return false
	}
	s := &ReadPartSelector{Page: &page.Page, Limit: &page.Limit, ParentCommentID: page.ParentCommentID}
	return validDeferredPart(page.ContentPage, page.ResourceID, page.ResourceKind, "comments_page", s)
}

func validDeferredBoard(board *ReadMeetingBoard, q MeetingReadRequest) bool {
	if board.Complete == nil || *board.Complete || board.NextCursor != nil || len(board.Cards) != 0 || len(board.Groups) != 0 {
		return false
	}
	s := &ReadPartSelector{Limit: &q.BoardLimit}
	if q.BoardCursor != "" {
		s.BoardCursor = &q.BoardCursor
	}
	return validDeferredPart(board.ContentPage, q.MeetingID, "meeting", "board_page", s)
}
