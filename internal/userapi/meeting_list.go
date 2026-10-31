package userapi

import "context"

const (
	meetingListQuery    = `query RiidoCLIMeetingList($componentId:String!,$teamId:String!){getComponentReadScope(componentId:$componentId,teamId:$teamId){schemaVersion rootId teamId status selection{relatedMeetingLimit} meetings{id teamId reasons{kind relation source}}}}`
	meetingSummaryQuery = `query RiidoCLIMeetingSummary($meetingId:String!,$teamId:String!){getMeetingReadScope(meetingId:$meetingId,teamId:$teamId,boardLimit:1){id teamId status summary{id kind title}}}`
)

// MeetingList is the server's authorized one-hop selection, not a complete list
// of every linked meeting. Summary reads are live, sequential reauthorizations.
type MeetingList struct {
	WorkID         string           `json:"workId"`
	TeamID         string           `json:"teamId"`
	ContextStatus  string           `json:"contextStatus"`
	SelectionLimit int              `json:"selectionLimit"`
	MaxDepth       int              `json:"maxDepth"`
	Meetings       []MeetingListRow `json:"meetings"`
}

type MeetingListRow struct {
	ID      string             `json:"id"`
	TeamID  string             `json:"teamId"`
	Title   string             `json:"title"`
	Status  string             `json:"status"`
	Reasons []*ReadScopeReason `json:"reasons"`
}

type meetingListScope struct {
	SchemaVersion string                  `json:"schemaVersion"`
	RootID        string                  `json:"rootId"`
	TeamID        string                  `json:"teamId"`
	Status        string                  `json:"status"`
	Selection     *ReadScopeSelection     `json:"selection"`
	Meetings      []*ReadMeetingReference `json:"meetings"`
}

func (r Reader) Meetings(ctx context.Context, workID, teamID string) (*MeetingList, error) {
	if workID == "" || teamID == "" {
		return nil, ErrInvalidResponse
	}
	var data struct {
		Scope *meetingListScope `json:"getComponentReadScope"`
	}
	if err := r.read(ctx, meetingListQuery, map[string]string{"componentId": workID, "teamId": teamID}, &data); err != nil {
		return nil, err
	}
	if !validMeetingListScope(data.Scope, workID, teamID) {
		return nil, ErrInvalidResponse
	}
	result := &MeetingList{WorkID: workID, TeamID: teamID, ContextStatus: data.Scope.Status, SelectionLimit: data.Scope.Selection.RelatedMeetingLimit, MaxDepth: 1, Meetings: []MeetingListRow{}}
	for _, ref := range data.Scope.Meetings {
		row, err := r.meetingSummary(ctx, ref)
		if err != nil {
			return nil, err
		}
		result.Meetings = append(result.Meetings, row)
	}
	return result, nil
}

func validMeetingListScope(scope *meetingListScope, workID, teamID string) bool {
	if scope == nil || scope.SchemaVersion != ReadScopeSchemaVersion || scope.RootID != workID || scope.TeamID != teamID || scope.Selection == nil || scope.Selection.RelatedMeetingLimit != 5 || scope.Meetings == nil || len(scope.Meetings) > 5 || (scope.Status != "available" && scope.Status != "partial") {
		return false
	}
	seen := map[string]bool{}
	for _, ref := range scope.Meetings {
		if ref == nil || ref.ID == "" || ref.TeamID == "" || seen[ref.ID] || !validScopeReasons(ref.Reasons) {
			return false
		}
		seen[ref.ID] = true
		for _, reason := range ref.Reasons {
			if reason.Kind != "meeting" || reason.Relation != nil || reason.Source != nil {
				return false
			}
		}
	}
	return true
}

func (r Reader) meetingSummary(ctx context.Context, ref *ReadMeetingReference) (MeetingListRow, error) {
	var data struct {
		Meeting *struct {
			ID, TeamID, Status string
			Summary            *struct {
				ID, Kind string
				Title    *string
			}
		} `json:"getMeetingReadScope"`
	}
	if err := r.read(ctx, meetingSummaryQuery, map[string]string{"meetingId": ref.ID, "teamId": ref.TeamID}, &data); err != nil {
		return MeetingListRow{}, err
	}
	m := data.Meeting
	if m == nil || m.ID != ref.ID || m.TeamID != ref.TeamID || (m.Status != "available" && m.Status != "partial") || m.Summary == nil || m.Summary.ID != ref.ID || m.Summary.Kind != "meeting" || m.Summary.Title == nil {
		return MeetingListRow{}, ErrInvalidResponse
	}
	return MeetingListRow{ID: ref.ID, TeamID: ref.TeamID, Title: *m.Summary.Title, Status: m.Status, Reasons: ref.Reasons}, nil
}
