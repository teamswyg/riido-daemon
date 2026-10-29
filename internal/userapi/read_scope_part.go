package userapi

import (
	"encoding/json"
	"math"
	"strings"
)

// This is a closed descriptor for a product page, never a query or authority.
type ReadPartSelector struct {
	ParentCommentID *string `json:"parentCommentId,omitempty"`
	Page            *int    `json:"page,omitempty"`
	Limit           *int    `json:"limit,omitempty"`
	BoardCursor     *string `json:"boardCursor,omitempty"`
}

func normalizedReadPartSelector(kind, part string, in *ReadPartSelector) (*ReadPartSelector, bool) {
	s := ReadPartSelector{}
	if in != nil {
		s = *in
	}
	empty := s.ParentCommentID == nil && s.Page == nil && s.Limit == nil && s.BoardCursor == nil
	switch part {
	case "document", "description", "document_alternate":
		return nil, empty
	case "metadata":
		return &ReadPartSelector{}, empty
	case "comments_page":
		if kind == "meeting" || s.BoardCursor != nil {
			return nil, false
		}
		page, limit := 1, 20
		if s.Page != nil {
			page = *s.Page
		}
		if s.Limit != nil {
			limit = *s.Limit
		}
		if page < 1 || page > math.MaxInt32 || limit < 1 || limit > 100 || int64(page-1)*int64(limit) > math.MaxInt32 {
			return nil, false
		}
		s.Page, s.Limit = &page, &limit
		if s.ParentCommentID != nil {
			if *s.ParentCommentID == "" {
				s.ParentCommentID = nil
			} else if !validReadPartToken(*s.ParentCommentID, 256) {
				return nil, false
			}
		}
		return &s, true
	case "board_page":
		if kind != "meeting" || s.ParentCommentID != nil || s.Page != nil {
			return nil, false
		}
		limit := 100
		if s.Limit != nil {
			limit = *s.Limit
		}
		if limit < 1 || limit > 200 {
			return nil, false
		}
		s.Limit = &limit
		if s.BoardCursor != nil {
			if *s.BoardCursor == "" {
				s.BoardCursor = nil
			} else if !validReadPartToken(*s.BoardCursor, 1024) {
				return nil, false
			}
		}
		return &s, true
	default:
		return nil, false
	}
}

func validReadPartToken(value string, maxBytes int) bool {
	return len(value) <= maxBytes && strings.TrimSpace(value) == value && !strings.ContainsAny(value, "\x00\r\n")
}

func validReadPartSelector(page *ReadBodyPage, q ReadBodyRequest) bool {
	want, ok := normalizedReadPartSelector(q.Kind, q.Part, q.Selector)
	if !ok {
		return false
	}
	if want == nil {
		_, ok = normalizedReadPartSelector(q.Kind, q.Part, page.Selector)
		return ok
	}
	if page.Selector == nil || page.Source != "riido_"+q.Part || page.Format != "json" {
		return false
	}
	got, ok := normalizedReadPartSelector(q.Kind, q.Part, page.Selector)
	if !ok {
		return false
	}
	a, _ := json.Marshal(want)
	b, _ := json.Marshal(got)
	return string(a) == string(b)
}
