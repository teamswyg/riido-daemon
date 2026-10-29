package userapi

import (
	"context"
	"regexp"
	"unicode/utf8"
)

const (
	DefaultReadBodyLimit = 16384
	MaxReadBodyLimit     = 65536
)

var readBodyVersionPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type ReadBodyRequest struct {
	ComponentID string            `json:"componentId"`
	TeamID      string            `json:"teamId"`
	Part        string            `json:"part"`
	Version     string            `json:"version"`
	Offset      int               `json:"offset"`
	Limit       int               `json:"limit"`
	Kind        string            `json:"-"`
	Selector    *ReadPartSelector `json:"selector,omitempty"`
}

func (q ReadBodyRequest) Valid() bool {
	_, selectorOK := normalizedReadPartSelector(q.Kind, q.Part, q.Selector)
	return selectorOK && validBodyResourceKind(q.Kind) && q.ComponentID != "" && q.TeamID != "" &&
		readBodyVersionPattern.MatchString(q.Version) && q.Offset >= 0 && q.Offset <= 2147483647 &&
		q.Limit > 0 && q.Limit <= MaxReadBodyLimit
}

func (r Reader) Body(ctx context.Context, q ReadBodyRequest) (ReadBodyPage, error) {
	if !q.Valid() {
		return ReadBodyPage{}, ErrInvalidResponse
	}
	query, field, idKey := readBodyOperation(q.Kind)
	variables := map[string]any{idKey: q.ComponentID, "teamId": q.TeamID, "part": q.Part, "version": q.Version, "offset": q.Offset, "limit": q.Limit}
	selector, _ := normalizedReadPartSelector(q.Kind, q.Part, q.Selector)
	if selector != nil {
		variables["selector"] = selector
	}
	var data map[string]*ReadBodyPage
	if err := r.read(ctx, query, variables, &data); err != nil {
		return ReadBodyPage{}, err
	}
	if !validReadBodyPage(data[field], q) {
		return ReadBodyPage{}, ErrInvalidResponse
	}
	return *data[field], nil
}

func validReadBodyPage(page *ReadBodyPage, q ReadBodyRequest) bool {
	if page == nil || page.ResourceID != q.ComponentID || page.Part != q.Part || page.Version != q.Version || page.Offset == nil || *page.Offset != q.Offset || page.Complete == nil {
		return false
	}
	if page.Source == "" || page.Format == "" || !utf8.ValidString(page.Content) || len(page.Content) > q.Limit {
		return false
	}
	if !validReadPartSelector(page, q) {
		return false
	}
	if *page.Complete {
		return page.NextOffset == nil && ((page.Status == "empty" && page.Content == "") || page.Status == "available")
	}
	return page.Status == "partial" && len(page.Content) > 0 && page.NextOffset != nil &&
		*page.NextOffset > q.Offset && int64(*page.NextOffset) == int64(q.Offset)+int64(len(page.Content))
}
