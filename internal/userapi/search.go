package userapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const workSearchQuery = `query RiidoCLIWorkSearch($workspaceId:String!,$query:String!,$cursor:String,$limit:Int){searchWorkspaceWork(workspaceId:$workspaceId,query:$query,cursor:$cursor,limit:$limit){schemaVersion workspaceId query limit complete nextCursor items{kind id teamId title componentType key keyNumber updatedAt matchedFields}}}`

type WorkSearchRequest struct {
	WorkspaceID, Query, Cursor string
	Limit                      int
}

func (q WorkSearchRequest) NormalizedQuery() string {
	return strings.ToLower(strings.Join(strings.Fields(q.Query), " "))
}

func (q WorkSearchRequest) Valid() bool {
	return q.WorkspaceID != "" && len(q.WorkspaceID) <= 200 && strings.TrimSpace(q.WorkspaceID) == q.WorkspaceID && utf8.ValidString(q.WorkspaceID) && !strings.ContainsAny(q.WorkspaceID, "\x00\r\n") && utf8.ValidString(q.Query) && len(q.Query) <= 4096 && !strings.ContainsRune(q.Query, 0) && q.NormalizedQuery() != "" && utf8.RuneCountInString(q.NormalizedQuery()) <= 200 && len(q.Cursor) <= 2048 && utf8.ValidString(q.Cursor) && !strings.ContainsRune(q.Cursor, 0) && q.Limit >= 0 && q.Limit <= 100
}

type WorkSearchItem struct {
	Kind          string   `json:"kind"`
	ID            string   `json:"id"`
	TeamID        string   `json:"teamId"`
	Title         *string  `json:"title"`
	ComponentType *string  `json:"componentType"`
	Key           *string  `json:"key"`
	KeyNumber     *string  `json:"keyNumber"`
	UpdatedAt     string   `json:"updatedAt"`
	MatchedFields []string `json:"matchedFields"`
}
type WorkSearchPage struct {
	SchemaVersion string            `json:"schemaVersion"`
	WorkspaceID   string            `json:"workspaceId"`
	Query         string            `json:"query"`
	Limit         int               `json:"limit"`
	Items         []*WorkSearchItem `json:"items"`
	NextCursor    *string           `json:"nextCursor"`
	Complete      bool              `json:"complete"`
}

func (r Reader) Search(ctx context.Context, q WorkSearchRequest) (*WorkSearchPage, error) {
	if !q.Valid() {
		return nil, ErrInvalidResponse
	}
	if q.Limit == 0 {
		q.Limit = 20
	}
	var cursor *string
	if q.Cursor != "" {
		cursor = &q.Cursor
	}
	var data struct {
		Page json.RawMessage `json:"searchWorkspaceWork"`
	}
	if err := r.read(ctx, workSearchQuery, map[string]any{"workspaceId": q.WorkspaceID, "query": q.NormalizedQuery(), "cursor": cursor, "limit": q.Limit}, &data); err != nil {
		return nil, err
	}
	var page struct {
		WorkSearchPage
		Complete *bool `json:"complete"`
	}
	if json.Unmarshal(data.Page, &page) != nil || page.Complete == nil || !workSearchNullableFieldsPresent(data.Page) {
		return nil, ErrInvalidResponse
	}
	p := &page.WorkSearchPage
	p.Complete = *page.Complete
	if p.SchemaVersion != "riido-work-search.v1" || p.WorkspaceID != q.WorkspaceID || p.Query != q.NormalizedQuery() || p.Limit != q.Limit || p.Items == nil || len(p.Items) > q.Limit {
		return nil, ErrInvalidResponse
	}
	if p.Complete {
		if p.NextCursor != nil {
			return nil, ErrInvalidResponse
		}
	} else if len(p.Items) != p.Limit || p.NextCursor == nil || *p.NextCursor == "" || len(*p.NextCursor) > 2048 || *p.NextCursor == q.Cursor {
		return nil, ErrInvalidResponse
	}
	if p.NextCursor != nil {
		decoded, err := base64.RawURLEncoding.Strict().DecodeString(*p.NextCursor)
		if err != nil || base64.RawURLEncoding.EncodeToString(decoded) != *p.NextCursor {
			return nil, ErrInvalidResponse
		}
	}
	seen := map[string]bool{}
	var priorTime time.Time
	priorKind, priorID := "", ""
	for _, item := range p.Items {
		if item == nil || item.ID == "" || item.TeamID == "" || item.Title == nil || !slices.Contains([]string{"component", "backlog", "intake"}, item.Kind) || seen[item.Kind+":"+item.ID] || len(item.MatchedFields) == 0 {
			return nil, ErrInvalidResponse
		}
		if item.Kind == "component" {
			if item.ComponentType == nil || *item.ComponentType == "" || item.Key == nil || item.KeyNumber == nil {
				return nil, ErrInvalidResponse
			}
		} else if item.ComponentType != nil || item.Key != nil || item.KeyNumber != nil {
			return nil, ErrInvalidResponse
		}
		stamp, err := time.Parse(time.RFC3339Nano, item.UpdatedAt)
		if err != nil {
			return nil, ErrInvalidResponse
		}
		if !priorTime.IsZero() && (stamp.After(priorTime) || stamp.Equal(priorTime) && (item.Kind < priorKind || item.Kind == priorKind && item.ID <= priorID)) {
			return nil, ErrInvalidResponse
		}
		fields := map[string]bool{}
		for _, field := range item.MatchedFields {
			if !slices.Contains([]string{"title", "key", "description", "document"}, field) || fields[field] || field == "key" && item.Key == nil {
				return nil, ErrInvalidResponse
			}
			fields[field] = true
		}
		seen[item.Kind+":"+item.ID] = true
		priorTime, priorKind, priorID = stamp, item.Kind, item.ID
	}
	return p, nil
}

// GraphQL nullable selections must be present with an explicit null. Missing
// fields indicate an incomplete response, not a completed empty page.
func workSearchNullableFieldsPresent(raw json.RawMessage) bool {
	var fields struct {
		NextCursor json.RawMessage              `json:"nextCursor"`
		Items      []map[string]json.RawMessage `json:"items"`
	}
	if json.Unmarshal(raw, &fields) != nil || len(fields.NextCursor) == 0 {
		return false
	}
	for _, item := range fields.Items {
		for _, name := range []string{"componentType", "key", "keyNumber"} {
			if len(item[name]) == 0 {
				return false
			}
		}
	}
	return true
}
