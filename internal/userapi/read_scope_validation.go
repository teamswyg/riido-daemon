package userapi

import "strings"

func validScopeReasons(reasons []*ReadScopeReason) bool {
	if len(reasons) == 0 {
		return false
	}
	for _, reason := range reasons {
		if reason == nil || strings.TrimSpace(reason.Kind) == "" {
			return false
		}
		if reason.Source != nil && *reason.Source != "document" && *reason.Source != "document_alternate" {
			return false
		}
	}
	return true
}

func validInitialReadBody(body *ReadBodyPage, id, part string) bool {
	if body == nil || body.ResourceID != id || body.Part != part || body.Complete == nil || body.Offset == nil || *body.Offset != 0 {
		return false
	}
	if body.Status == "unavailable" || body.Status == "denied" || body.Status == "unsupported" {
		return !*body.Complete && body.Content == "" && body.NextOffset == nil
	}
	q := ReadBodyRequest{ComponentID: id, TeamID: "scope-authorized", Part: part, Version: body.Version, Limit: MaxReadBodyLimit}
	return q.Valid() && validReadBodyPage(body, q)
}

func validScopeComments(page *ReadScopeCommentPage) bool {
	if page == nil || page.Items == nil || page.Complete == nil || page.Page < 1 {
		return false
	}
	if page.ContentPage != nil {
		return validDeferredComments(page)
	}
	if *page.Complete {
		return page.NextPage == nil && (page.Status == "available" || page.Status == "empty")
	}
	if page.Status != "partial" && page.Status != "unavailable" && page.Status != "denied" && page.Status != "unsupported" {
		return false
	}
	return page.NextPage == nil || (page.Status == "partial" && *page.NextPage > page.Page)
}

func validInitialScopeComments(page *ReadScopeCommentPage, id, kind string) bool {
	if !validScopeComments(page) || page.ResourceID != id || page.ResourceKind != kind || page.ParentCommentID != nil || page.Page != 1 || page.Limit < 1 || page.Limit > 100 {
		return false
	}
	if page.Status == "unavailable" || page.Status == "denied" || page.Status == "unsupported" {
		return !*page.Complete && len(page.Items) == 0 && page.NextPage == nil
	}
	return validReadScopeCommentPage(page, ReadScopeCommentsRequest{ResourceID: id, ResourceKind: kind, TeamID: "scope-authorized", Page: 1, Limit: page.Limit})
}

func incompleteReadParts(parts []*ReadPartStatus) bool {
	for _, part := range parts {
		if !*part.Complete {
			return true
		}
	}
	return false
}
