package userapi

import (
	"context"
	"math"
)

type ReadScopeCommentsRequest struct {
	ResourceID      string `json:"resourceId"`
	ResourceKind    string `json:"resourceKind"`
	TeamID          string `json:"teamId"`
	Page            int    `json:"page"`
	Limit           int    `json:"limit"`
	ParentCommentID string `json:"parentCommentId,omitempty"`
	Version         string `json:"version,omitempty"`
}

func (q ReadScopeCommentsRequest) Valid() bool {
	kind := q.ResourceKind == "component" || q.ResourceKind == "backlog" || q.ResourceKind == "intake" || q.ResourceKind == "meeting"
	return kind && q.ResourceID != "" && q.TeamID != "" && q.Page > 0 && q.Page <= math.MaxInt32 && q.Limit > 0 && q.Limit <= 100 &&
		int64(q.Page-1)*int64(q.Limit) <= math.MaxInt32 && (q.Version != "" || q.Page == 1) &&
		(q.Version == "" || readBodyVersionPattern.MatchString(q.Version))
}

func (r Reader) ScopeComments(ctx context.Context, q ReadScopeCommentsRequest) (ReadScopeCommentPage, error) {
	if !q.Valid() {
		return ReadScopeCommentPage{}, ErrInvalidResponse
	}
	var data struct {
		Page *ReadScopeCommentPage `json:"getReadScopeComments"`
	}
	if err := r.read(ctx, readScopeCommentsQuery, q, &data); err != nil {
		return ReadScopeCommentPage{}, err
	}
	if !validReadScopeCommentPage(data.Page, q) {
		return ReadScopeCommentPage{}, ErrInvalidResponse
	}
	return *data.Page, nil
}

func validReadScopeCommentPage(page *ReadScopeCommentPage, q ReadScopeCommentsRequest) bool {
	if !validScopeComments(page) || page.ResourceID != q.ResourceID || page.ResourceKind != q.ResourceKind || page.Page != q.Page || page.Limit != q.Limit || len(page.Items) > q.Limit || !readBodyVersionPattern.MatchString(page.Version) {
		return false
	}
	if q.Version != "" && page.Version != q.Version {
		return false
	}
	if (q.ParentCommentID == "" && page.ParentCommentID != nil) || (q.ParentCommentID != "" && (page.ParentCommentID == nil || *page.ParentCommentID != q.ParentCommentID)) {
		return false
	}
	return page.ContentPage != nil || *page.Complete || (page.Status == "partial" && page.NextPage != nil && *page.NextPage == page.Page+1 && len(page.Items) > 0)
}
