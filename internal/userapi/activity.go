package userapi

import (
	"context"
	"strconv"
)

const activityQuery = `query RiidoCLIActivity($componentId:String!,$teamId:String!,$pagination:ComponentP06ActivityPaginationInput!){getComponentActivityLogs(componentId:$componentId,teamId:$teamId,offsetBasedPaginationDto:$pagination){activityLogs{id createdAt category action property oldValue newValue author{id status workspaceProfile{nickname status avatar{fileLocation}}}}meta{total currentPage nextPage prevPage lastPage}}}`

// ActivityPagination retains the existing server's String input contract.
type ActivityPagination CommentPagination

func (p ActivityPagination) Valid() bool {
	page, limit, valid := CommentPagination(p).valid()
	return valid && p.Page == strconv.FormatInt(page, 10) && p.Limit == strconv.FormatInt(limit, 10)
}

type ActivityLog struct {
	ID        string          `json:"id"`
	CreatedAt *string         `json:"createdAt"`
	Category  *string         `json:"category"`
	Action    *string         `json:"action"`
	Property  *string         `json:"property"`
	OldValue  *string         `json:"oldValue"`
	NewValue  *string         `json:"newValue"`
	Author    *ActivityAuthor `json:"author"`
}

type ActivityAuthor struct {
	ID               string           `json:"id"`
	Status           *string          `json:"status"`
	WorkspaceProfile *ActivityProfile `json:"workspaceProfile"`
}

type ActivityProfile struct {
	Nickname *string         `json:"nickname"`
	Status   *string         `json:"status"`
	Avatar   *ActivityAvatar `json:"avatar"`
}

type ActivityAvatar struct {
	FileLocation *string `json:"fileLocation"`
}

type ActivityPage struct {
	ActivityLogs []*ActivityLog `json:"activityLogs"`
	Meta         *OffsetMeta    `json:"meta"`
}

// Activity is one live authorized page, not a snapshot or a recursive read.
func (r Reader) Activity(ctx context.Context, id, teamID string, pagination ActivityPagination) (ActivityPage, error) {
	if id == "" || teamID == "" || !pagination.Valid() {
		return ActivityPage{}, ErrInvalidResponse
	}
	var data struct {
		Page *ActivityPage `json:"getComponentActivityLogs"`
	}
	variables := struct {
		ComponentID string             `json:"componentId"`
		TeamID      string             `json:"teamId"`
		Pagination  ActivityPagination `json:"pagination"`
	}{id, teamID, pagination}
	if err := r.read(ctx, activityQuery, variables, &data); err != nil {
		return ActivityPage{}, err
	}
	if data.Page == nil || !validActivityLogs(data.Page.ActivityLogs) || !validCommentMeta(data.Page.Meta, CommentPagination(pagination)) {
		return ActivityPage{}, ErrInvalidResponse
	}
	page, limit, _ := CommentPagination(pagination).valid()
	remaining := max(int64(*data.Page.Meta.Total)-(page-1)*limit, 0)
	if int64(len(data.Page.ActivityLogs)) != min(remaining, limit) {
		return ActivityPage{}, ErrInvalidResponse
	}
	return *data.Page, nil
}

func validActivityLogs(logs []*ActivityLog) bool {
	if logs == nil {
		return false
	}
	for _, entry := range logs {
		if entry == nil || entry.ID == "" || entry.CreatedAt == nil || entry.Category == nil || entry.Action == nil {
			return false
		}
		if author := entry.Author; author != nil {
			profile := author.WorkspaceProfile
			if author.ID == "" || author.Status == nil || profile == nil || profile.Nickname == nil || profile.Status == nil {
				return false
			}
		}
	}
	return true
}
