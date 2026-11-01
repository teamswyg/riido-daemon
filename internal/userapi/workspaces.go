package userapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"regexp"
)

const workspaceListQuery = `query RiidoCLIWorkspaces($cursor:String,$limit:Int){getCLIWorkspaces(cursor:$cursor,limit:$limit){schemaVersion userId limit items{id name isDemo} nextCursor complete}}`

var workspaceIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

type WorkspaceListRequest struct {
	Cursor *string `json:"cursor,omitempty"`
	Limit  int     `json:"limit"`
}

// Valid checks the public bounds. Cursor authorization remains server-owned.
func (q WorkspaceListRequest) Valid() bool {
	return q.Limit >= 1 && q.Limit <= 100 && (q.Cursor == nil || validWorkspaceCursor(*q.Cursor))
}

func validWorkspaceCursor(cursor string) bool {
	if cursor == "" || len(cursor) > 1024 {
		return false
	}
	raw, err := base64.RawURLEncoding.Strict().DecodeString(cursor)
	return err == nil && base64.RawURLEncoding.EncodeToString(raw) == cursor
}

type WorkspaceList struct {
	SchemaVersion string          `json:"schemaVersion"`
	UserID        string          `json:"userId"`
	Limit         int             `json:"limit"`
	Items         []WorkspaceItem `json:"items"`
	NextCursor    *string         `json:"nextCursor"`
	Complete      bool            `json:"complete"`
}

type WorkspaceItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	IsDemo bool   `json:"isDemo"`
}

func (r Reader) Workspaces(ctx context.Context, q WorkspaceListRequest) (*WorkspaceList, error) {
	if !q.Valid() {
		return nil, ErrInvalidRequest
	}
	var data struct {
		Page *struct {
			SchemaVersion string
			UserID        string
			Limit         int
			Items         []*struct {
				ID     string
				Name   *string
				IsDemo *bool
			}
			NextCursor json.RawMessage
			Complete   *bool
		} `json:"getCLIWorkspaces"`
	}
	if err := r.read(ctx, workspaceListQuery, q, &data); err != nil {
		return nil, err
	}
	p := data.Page
	if p == nil || p.SchemaVersion != "riido-cli-workspaces.v1" || !workspaceIDPattern.MatchString(p.UserID) || p.Limit != q.Limit || p.Items == nil || len(p.Items) > q.Limit || p.Complete == nil {
		return nil, ErrInvalidResponse
	}
	var nextCursor *string
	if len(p.NextCursor) == 0 || json.Unmarshal(p.NextCursor, &nextCursor) != nil {
		return nil, ErrInvalidResponse
	}
	if *p.Complete {
		if nextCursor != nil {
			return nil, ErrInvalidResponse
		}
	} else if len(p.Items) != q.Limit || nextCursor == nil || !validWorkspaceCursor(*nextCursor) || (q.Cursor != nil && *q.Cursor == *nextCursor) {
		return nil, ErrInvalidResponse
	}
	result := &WorkspaceList{SchemaVersion: p.SchemaVersion, UserID: p.UserID, Limit: p.Limit, Items: []WorkspaceItem{}, NextCursor: nextCursor, Complete: *p.Complete}
	previous := ""
	for _, item := range p.Items {
		if item == nil || !workspaceIDPattern.MatchString(item.ID) || item.ID <= previous || item.Name == nil || item.IsDemo == nil {
			return nil, ErrInvalidResponse
		}
		result.Items = append(result.Items, WorkspaceItem{ID: item.ID, Name: *item.Name, IsDemo: *item.IsDemo})
		previous = item.ID
	}
	return result, nil
}
