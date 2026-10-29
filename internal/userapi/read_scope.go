package userapi

import (
	"context"
	"encoding/json"
)

func (r Reader) Context(ctx context.Context, id, teamID string) (ReadScope, error) {
	if id == "" || teamID == "" {
		return ReadScope{}, ErrInvalidResponse
	}
	var data struct {
		Scope *ReadScope `json:"getComponentReadScope"`
	}
	if err := r.read(ctx, readScopeQuery, map[string]string{"componentId": id, "teamId": teamID}, &data); err != nil {
		return ReadScope{}, err
	}
	if !validReadScope(data.Scope, id, teamID) {
		return ReadScope{}, ErrInvalidResponse
	}
	for _, work := range data.Scope.Works {
		if work.ContentPage != nil || !*work.Body.Complete || !*work.Comments.Complete || incompleteReadParts(work.PartsStatus) ||
			(work.DescriptionBody != nil && !*work.DescriptionBody.Complete) || (work.AlternateBody != nil && !*work.AlternateBody.Complete) {
			data.Scope.Status = "partial"
		}
	}
	// Resolve only the references selected by the root collector. The meeting
	// endpoint returns detail and board pages, never another relation traversal.
	for _, ref := range data.Scope.Meetings {
		meeting, err := r.Meeting(ctx, MeetingReadRequest{MeetingID: ref.ID, TeamID: ref.TeamID, BoardLimit: 100})
		if err != nil {
			return ReadScope{}, err
		}
		ref.Snapshot = &meeting
		if meeting.Status == "partial" || !*meeting.Board.Complete {
			data.Scope.Status = "partial"
		}
	}
	return *data.Scope, nil
}

func validReadScope(scope *ReadScope, id, team string) bool {
	if scope == nil || scope.SchemaVersion != ReadScopeSchemaVersion || scope.RootID != id || scope.TeamID != team || scope.Selection == nil || scope.Works == nil || scope.Meetings == nil {
		return false
	}
	if scope.Status != "available" && scope.Status != "partial" {
		return false
	}
	if scope.Selection.RelatedWorkLimit < 0 || scope.Selection.RelatedMeetingLimit < 0 || scope.Selection.ChildrenLimit < 0 {
		return false
	}
	seen, root := map[string]bool{}, false
	for _, work := range scope.Works {
		if work == nil || work.ID == "" || work.TeamID == "" || seen[work.Kind+":"+work.ID] || !validScopeReasons(work.Reasons) {
			return false
		}
		if work.Kind != "component" && work.Kind != "backlog" && work.Kind != "intake" {
			return false
		}
		seen[work.Kind+":"+work.ID] = true
		if work.ID == id && work.Kind == "component" && work.TeamID == team {
			root = true
		}
		if !validReadScopeWork(work) {
			return false
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

func validReadScopeWork(work *ReadScopeWork) bool {
	var detail struct {
		ID string `json:"id"`
	}
	raw := work.Details
	if work.Kind == "backlog" {
		raw = work.BacklogDetails
	}
	if work.Kind == "intake" {
		raw = work.IntakeDetails
	}
	detailOK := json.Unmarshal(raw, &detail) == nil && detail.ID == work.ID
	if work.ContentPage != nil {
		detailOK = validReadSummary(work.Summary, work.ID, work.Kind) && validDeferredPart(work.ContentPage, work.ID, work.Kind, "metadata", nil)
	}
	return detailOK &&
		validInitialReadBody(work.Body, work.ID, "document") &&
		(work.DescriptionBody == nil || validInitialReadBody(work.DescriptionBody, work.ID, "description")) &&
		(work.AlternateBody == nil || validInitialReadBody(work.AlternateBody, work.ID, "document_alternate")) &&
		validInitialScopeComments(work.Comments, work.ID, work.Kind) &&
		work.Links != nil && work.GitHub != nil && work.Repositories != nil && work.Sprints != nil && work.Priorities != nil && work.Checklists != nil && work.Attachments != nil && validReadPartStatuses(work.PartsStatus)
}
