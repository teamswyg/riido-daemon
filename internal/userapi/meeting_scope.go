package userapi

import (
	"context"
	"encoding/json"
	"strings"
)

func (r Reader) Meeting(ctx context.Context, q MeetingReadRequest) (ReadMeeting, error) {
	if !q.Valid() {
		return ReadMeeting{}, ErrInvalidResponse
	}
	var data struct {
		Meeting *ReadMeeting `json:"getMeetingReadScope"`
	}
	if err := r.read(ctx, meetingScopeQuery, q, &data); err != nil {
		return ReadMeeting{}, err
	}
	if !validReadMeeting(data.Meeting, q) {
		return ReadMeeting{}, ErrInvalidResponse
	}
	if data.Meeting.ContentPage != nil || !*data.Meeting.Board.Complete || !*data.Meeting.Body.Complete || !*data.Meeting.DescriptionBody.Complete ||
		incompleteReadParts(data.Meeting.PartsStatus) || (data.Meeting.AlternateBody != nil && !*data.Meeting.AlternateBody.Complete) {
		data.Meeting.Status = "partial"
	}
	return *data.Meeting, nil
}

func validReadMeeting(meeting *ReadMeeting, q MeetingReadRequest) bool {
	if meeting == nil || meeting.ID != q.MeetingID || meeting.TeamID != q.TeamID || meeting.Labels == nil || !validReadPartStatuses(meeting.PartsStatus) {
		return false
	}
	if meeting.Status != "available" && meeting.Status != "partial" {
		return false
	}
	var details struct {
		ID           string            `json:"id"`
		Participants []json.RawMessage `json:"meetingParticipants"`
		Attendees    []json.RawMessage `json:"meetingAttendees"`
	}
	detailOK := json.Unmarshal(meeting.Details, &details) == nil && details.ID == q.MeetingID && details.Participants != nil && details.Attendees != nil
	if meeting.ContentPage != nil {
		detailOK = validReadSummary(meeting.Summary, q.MeetingID, "meeting") && validDeferredPart(meeting.ContentPage, q.MeetingID, "meeting", "metadata", nil)
	}
	if !detailOK {
		return false
	}
	if !validInitialReadBody(meeting.Body, q.MeetingID, "document") || !validInitialReadBody(meeting.DescriptionBody, q.MeetingID, "description") {
		return false
	}
	if meeting.AlternateBody != nil && !validInitialReadBody(meeting.AlternateBody, q.MeetingID, "document_alternate") {
		return false
	}
	b := meeting.Board
	if b == nil || b.Version == "" || b.Offset == nil || *b.Offset < 0 || b.Complete == nil || b.Groups == nil || b.Cards == nil || len(b.Cards) > q.BoardLimit {
		return false
	}
	if q.BoardCursor == "" && *b.Offset != 0 {
		return false
	}
	if b.ContentPage != nil {
		return validDeferredBoard(b, q)
	}
	if *b.Complete {
		return b.NextCursor == nil
	}
	return b.NextCursor != nil && strings.TrimSpace(*b.NextCursor) != "" && *b.NextCursor != q.BoardCursor && len(b.Cards) > 0
}

func validReadPartStatuses(parts []*ReadPartStatus) bool {
	if parts == nil {
		return false
	}
	seen := map[string]bool{}
	for _, part := range parts {
		if part == nil || part.Name == "" || part.Status == "" || part.Complete == nil || seen[part.Name] {
			return false
		}
		seen[part.Name] = true
	}
	return true
}
