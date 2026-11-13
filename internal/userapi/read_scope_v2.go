package userapi

import (
	"context"
	"encoding/json"
)

func (r Reader) WorkV2(ctx context.Context, locator WorkLocator) (WorkV2, error) {
	if !locator.valid() {
		return WorkV2{}, ErrInvalidRequest
	}
	var data struct {
		Work *WorkV2 `json:"getComponent"`
	}
	if err := r.read(ctx, workV2Query, locator.variables(), &data); err != nil {
		return WorkV2{}, err
	}
	w := data.Work
	if w == nil || w.SchemaVersion != "riido-work-detail.v2" || w.ID != locator.ID || w.WorkspaceID != locator.WorkspaceID ||
		w.ComponentType != locator.ComponentType || !sameLocatorTeam(w.OwnerTeamID, locator.TeamID) || w.CanEdit == nil || w.ComponentMembers == nil || w.ComponentLabels == nil {
		return WorkV2{}, ErrInvalidResponse
	}
	return *w, nil
}

func (r Reader) ContextV2(ctx context.Context, locator WorkLocator) (ReadScopeV2, error) {
	if !locator.valid() {
		return ReadScopeV2{}, ErrInvalidRequest
	}
	var data struct {
		Scope *ReadScopeV2 `json:"getComponentReadScope"`
	}
	if err := r.read(ctx, readScopeV2Query, locator.variables(), &data); err != nil {
		return ReadScopeV2{}, err
	}
	if !validReadScopeV2(data.Scope, locator) {
		return ReadScopeV2{}, ErrInvalidResponse
	}
	for _, work := range data.Scope.Works {
		if work.ContentPage != nil || !*work.Body.Complete || !*work.Comments.Complete || incompleteReadParts(work.PartsStatus) ||
			(work.DescriptionBody != nil && !*work.DescriptionBody.Complete) || (work.AlternateBody != nil && !*work.AlternateBody.Complete) {
			data.Scope.Status = "partial"
		}
	}
	// Meeting references retain their real teams and existing v1 reader.
	for _, ref := range data.Scope.Meetings {
		meeting, err := r.Meeting(ctx, MeetingReadRequest{MeetingID: ref.ID, TeamID: ref.TeamID, BoardLimit: 100})
		if err != nil {
			return ReadScopeV2{}, err
		}
		ref.Snapshot = &meeting
		if meeting.Status == "partial" || !*meeting.Board.Complete {
			data.Scope.Status = "partial"
		}
	}
	return *data.Scope, nil
}

func validReadScopeV2(scope *ReadScopeV2, locator WorkLocator) bool {
	if scope == nil || scope.SchemaVersion != "riido-read-scope.v2" || scope.RootID != locator.ID || scope.WorkspaceID != locator.WorkspaceID ||
		!sameLocatorTeam(scope.TeamID, locator.TeamID) || scope.Selection == nil || scope.Works == nil || scope.Meetings == nil ||
		(scope.Status != "available" && scope.Status != "partial") || scope.Selection.RelatedWorkLimit < 0 || scope.Selection.RelatedMeetingLimit < 0 || scope.Selection.ChildrenLimit < 0 {
		return false
	}
	seen, root := map[string]bool{}, false
	for _, work := range scope.Works {
		if work == nil || work.ID == "" || work.WorkspaceID != locator.WorkspaceID || seen[work.Kind+":"+work.ID] || !validScopeReasons(work.Reasons) ||
			(work.Kind != "component" && work.Kind != "backlog" && work.Kind != "intake") {
			return false
		}
		componentType := "task"
		if work.Kind == "component" {
			var detail struct {
				ComponentType string `json:"componentType"`
			}
			switch {
			case work.ContentPage != nil && work.Summary != nil && work.Summary.ComponentType != nil:
				componentType = *work.Summary.ComponentType
			case json.Unmarshal(work.Details, &detail) == nil:
				componentType = detail.ComponentType
			default:
				return false
			}
		}
		if componentType != "project" && componentType != "milestone" && componentType != "task" && componentType != "subtask" {
			return false
		}
		if !validLocatorTeam(2, work.Kind, componentType, work.TeamID) {
			return false
		}
		common := work.common()
		if !validReadScopeWork(&common) {
			return false
		}
		for _, body := range []*ReadBodyPageV2{work.Body, work.DescriptionBody, work.AlternateBody, work.ContentPage} {
			if body != nil && !validV2BodyLocator(body, work.WorkspaceID, work.TeamID, work.ID, work.Kind) {
				return false
			}
		}
		if !validV2CommentLocator(work.Comments, work.WorkspaceID, work.TeamID, work.ID, work.Kind) {
			return false
		}
		seen[work.Kind+":"+work.ID] = true
		if work.ID == locator.ID && sameLocatorTeam(work.TeamID, locator.TeamID) {
			if work.Kind == "component" && componentType != locator.ComponentType {
				return false
			}
			root = true
		}
	}
	for _, meeting := range scope.Meetings {
		if meeting == nil || meeting.ID == "" || meeting.TeamID == "" || seen["meeting:"+meeting.ID] || !validScopeReasons(meeting.Reasons) {
			return false
		}
		seen["meeting:"+meeting.ID] = true
	}
	return root
}

func validV2BodyLocator(page *ReadBodyPageV2, workspace string, team *string, id, kind string) bool {
	return page != nil && page.SchemaVersion == "riido-read-body.v2" && page.WorkspaceID == workspace && sameLocatorTeam(page.TeamID, team) && page.ResourceID == id && page.ResourceKind == kind
}

func validV2CommentLocator(page *ReadScopeCommentPageV2, workspace string, team *string, id, kind string) bool {
	return page != nil && page.SchemaVersion == "riido-read-comments.v2" && page.WorkspaceID == workspace && sameLocatorTeam(page.TeamID, team) && page.ResourceID == id && page.ResourceKind == kind &&
		(page.ContentPage == nil || validV2BodyLocator(page.ContentPage, workspace, team, id, kind))
}

func (r Reader) BodyV2(ctx context.Context, locator WorkLocator, q ReadBodyRequest) (ReadBodyPageV2, error) {
	if q.Kind == "" {
		q.Kind = "component"
	}
	if !locator.valid() || q.ComponentID != locator.ID || !q.ValidContent() || q.Kind == "meeting" || (q.Kind != "component" && locator.TeamID == nil) {
		return ReadBodyPageV2{}, ErrInvalidRequest
	}
	query, field, idKey := readScopeBodyV2Query, "getComponentReadScopeBody", "componentId"
	if q.Kind == "backlog" {
		query, field, idKey = readBacklogBodyV2Query, "getBacklogReadScopeBody", "backlogId"
	}
	if q.Kind == "intake" {
		query, field, idKey = readIntakeBodyV2Query, "getIntakeReadScopeBody", "intakeId"
	}
	variables := map[string]any{idKey: locator.ID, "workspaceId": locator.WorkspaceID, "teamId": locator.TeamID, "part": q.Part, "version": q.Version, "offset": q.Offset, "limit": q.Limit}
	if selector, _ := normalizedReadPartSelector(q.Kind, q.Part, q.Selector); selector != nil {
		variables["selector"] = selector
	}
	var data map[string]*ReadBodyPageV2
	if err := r.read(ctx, query, variables, &data); err != nil {
		return ReadBodyPageV2{}, err
	}
	page := data[field]
	if !validV2BodyLocator(page, locator.WorkspaceID, locator.TeamID, locator.ID, q.Kind) || !validReadBodyPage(&page.ReadBodyPage, q) {
		return ReadBodyPageV2{}, ErrInvalidResponse
	}
	return *page, nil
}

func (r Reader) ScopeCommentsV2(ctx context.Context, locator WorkLocator, q ReadScopeCommentsRequest) (ReadScopeCommentPageV2, error) {
	if !locator.valid() || q.ResourceID != locator.ID || !q.ValidPage() || q.ResourceKind == "meeting" || (q.ResourceKind != "component" && locator.TeamID == nil) {
		return ReadScopeCommentPageV2{}, ErrInvalidRequest
	}
	variables := map[string]any{"resourceId": locator.ID, "resourceKind": q.ResourceKind, "workspaceId": locator.WorkspaceID, "teamId": locator.TeamID, "page": q.Page, "limit": q.Limit}
	if q.ParentCommentID != "" {
		variables["parentCommentId"] = q.ParentCommentID
	}
	if q.Version != "" {
		variables["version"] = q.Version
	}
	var data struct {
		Page *ReadScopeCommentPageV2 `json:"getReadScopeComments"`
	}
	if err := r.read(ctx, readScopeCommentsV2Query, variables, &data); err != nil {
		return ReadScopeCommentPageV2{}, err
	}
	page := data.Page
	if !validV2CommentLocator(page, locator.WorkspaceID, locator.TeamID, locator.ID, q.ResourceKind) {
		return ReadScopeCommentPageV2{}, ErrInvalidResponse
	}
	common := page.common()
	if !validReadScopeCommentPage(&common, q) {
		return ReadScopeCommentPageV2{}, ErrInvalidResponse
	}
	return *page, nil
}
