package userapi

import (
	"context"
	"encoding/json"
	"math"
	"strconv"
)

const (
	commentFields     = `id status createdAt updatedAt body author{id status workspaceProfile{nickname status}} isFromSlack slackMessageUrl isFromDiscord discordMessageUrl`
	commentsQuery     = `query RiidoCLIComments($componentId:String!,$pagination:LegacyOffsetPaginationInput!){getComponentComments(componentId:$componentId,offsetBasedPaginationDto:$pagination){comments{` + commentFields + ` replyCount}meta{total currentPage nextPage prevPage lastPage}}}`
	repliesQuery      = `query RiidoCLIReplies($componentId:String!,$teamId:String!,$commentId:String!){getCommentReplies(componentId:$componentId,teamId:$teamId,commentId:$commentId){` + commentFields + `}}`
	commentScopeQuery = `query RiidoCLICommentScope($componentId:String!,$teamId:String!){getComponent(componentId:$componentId,teamId:$teamId,getComponentDto:{}){id}}`
)

type Comment struct {
	ID                string         `json:"id"`
	Status            string         `json:"status"`
	CreatedAt         string         `json:"createdAt"`
	UpdatedAt         string         `json:"updatedAt"`
	Body              string         `json:"body"`
	Author            *CommentAuthor `json:"author"`
	ReplyCount        *int32         `json:"replyCount,omitempty"`
	IsFromSlack       bool           `json:"isFromSlack"`
	SlackMessageURL   *string        `json:"slackMessageUrl"`
	IsFromDiscord     bool           `json:"isFromDiscord"`
	DiscordMessageURL *string        `json:"discordMessageUrl"`
}

type CommentAuthor struct {
	ID               string `json:"id"`
	Status           string `json:"status"`
	WorkspaceProfile struct {
		Nickname string `json:"nickname"`
		Status   string `json:"status"`
	} `json:"workspaceProfile"`
}

type CommentPage struct {
	Comments []*Comment  `json:"comments"`
	Meta     *OffsetMeta `json:"meta"`
}

// Required scalar pointers distinguish malformed missing/null metadata from 0.
type OffsetMeta struct {
	Total       *int32          `json:"total"`
	CurrentPage *int32          `json:"currentPage"`
	NextPage    json.RawMessage `json:"nextPage"`
	PrevPage    json.RawMessage `json:"prevPage"`
	LastPage    *int32          `json:"lastPage"`
}

type CommentPagination struct {
	Page  string `json:"page"`
	Limit string `json:"limit"`
}

func (p CommentPagination) valid() (int64, int64, bool) {
	page, pageErr := strconv.ParseInt(p.Page, 10, 32)
	limit, limitErr := strconv.ParseInt(p.Limit, 10, 32)
	valid := pageErr == nil && limitErr == nil && page > 0 && limit > 0 && limit <= 100 && (page-1)*limit <= math.MaxInt32
	return page, limit, valid
}

func (p CommentPagination) Valid() bool { _, _, valid := p.valid(); return valid }

// Comments performs a minimal team check before the independently authorized
// comments read. The two HTTP reads are not an authorization transaction.
func (r Reader) Comments(ctx context.Context, id, teamID string, pagination CommentPagination) (CommentPage, error) {
	if id == "" || teamID == "" || !pagination.Valid() {
		return CommentPage{}, ErrInvalidResponse
	}
	var scope struct {
		Component *struct {
			ID string `json:"id"`
		} `json:"getComponent"`
	}
	if err := r.read(ctx, commentScopeQuery, map[string]string{"componentId": id, "teamId": teamID}, &scope); err != nil {
		return CommentPage{}, err
	}
	if scope.Component == nil || scope.Component.ID != id {
		return CommentPage{}, ErrInvalidResponse
	}
	var data struct {
		Page *CommentPage `json:"getComponentComments"`
	}
	variables := struct {
		ComponentID string            `json:"componentId"`
		Pagination  CommentPagination `json:"pagination"`
	}{id, pagination}
	if err := r.read(ctx, commentsQuery, variables, &data); err != nil {
		return CommentPage{}, err
	}
	if data.Page == nil || !validComments(data.Page.Comments, true) || !validCommentMeta(data.Page.Meta, pagination) {
		return CommentPage{}, ErrInvalidResponse
	}
	_, limit, _ := pagination.valid()
	if int64(len(data.Page.Comments)) > limit {
		return CommentPage{}, ErrInvalidResponse
	}
	return *data.Page, nil
}

func (r Reader) Replies(ctx context.Context, id, teamID, commentID string) ([]*Comment, error) {
	if id == "" || teamID == "" || commentID == "" {
		return nil, ErrInvalidResponse
	}
	var data struct {
		Replies []*Comment `json:"getCommentReplies"`
	}
	if err := r.read(ctx, repliesQuery, map[string]string{"componentId": id, "teamId": teamID, "commentId": commentID}, &data); err != nil {
		return nil, err
	}
	if !validComments(data.Replies, false) {
		return nil, ErrInvalidResponse
	}
	return data.Replies, nil
}

func validComments(comments []*Comment, parents bool) bool {
	if comments == nil {
		return false
	}
	for _, comment := range comments {
		if comment == nil || comment.ID == "" || comment.Status == "" || comment.CreatedAt == "" || comment.UpdatedAt == "" || (parents && (comment.ReplyCount == nil || *comment.ReplyCount < 0)) {
			return false
		}
	}
	return true
}

func validCommentMeta(meta *OffsetMeta, pagination CommentPagination) bool {
	page, limit, valid := pagination.valid()
	if !valid || meta == nil || meta.Total == nil || *meta.Total < 0 || meta.CurrentPage == nil || int64(*meta.CurrentPage) != page || meta.LastPage == nil {
		return false
	}
	last := (int64(*meta.Total) + limit - 1) / limit
	if int64(*meta.LastPage) != last {
		return false
	}
	var next, prev *int64
	if page < last {
		value := page + 1
		next = &value
	}
	if page > 1 {
		value := page - 1
		prev = &value
	}
	return validPageNumber(meta.NextPage, next) && validPageNumber(meta.PrevPage, prev)
}

func validPageNumber(raw json.RawMessage, expected *int64) bool {
	if len(raw) == 0 {
		return false
	}
	var actual *int64
	if json.Unmarshal(raw, &actual) != nil {
		return false
	}
	if actual == nil || expected == nil {
		return actual == nil && expected == nil
	}
	return *actual == *expected
}
