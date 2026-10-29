package userapi

import "encoding/json"

type MeetingReadRequest struct {
	MeetingID   string `json:"meetingId"`
	TeamID      string `json:"teamId"`
	BoardCursor string `json:"boardCursor,omitempty"`
	BoardLimit  int    `json:"boardLimit"`
}

func (q MeetingReadRequest) Valid() bool {
	return q.MeetingID != "" && q.TeamID != "" && q.BoardLimit >= 1 && q.BoardLimit <= 200 && len(q.BoardCursor) <= 1024
}

type ReadMeeting struct {
	Summary         *ReadResourceSummary `json:"summary,omitempty"`
	ContentPage     *ReadBodyPage        `json:"contentPage,omitempty"`
	ID              string               `json:"id"`
	TeamID          string               `json:"teamId"`
	Status          string               `json:"status"`
	Details         json.RawMessage      `json:"details"`
	Labels          []json.RawMessage    `json:"labels"`
	Board           *ReadMeetingBoard    `json:"board"`
	Body            *ReadBodyPage        `json:"body"`
	DescriptionBody *ReadBodyPage        `json:"descriptionBody"`
	AlternateBody   *ReadBodyPage        `json:"alternateBody"`
	PartsStatus     []*ReadPartStatus    `json:"partsStatus"`
}

type ReadPartStatus struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Complete *bool  `json:"complete"`
}

type ReadMeetingBoard struct {
	ContentPage *ReadBodyPage     `json:"contentPage,omitempty"`
	Version     string            `json:"version"`
	Offset      *int              `json:"offset"`
	NextCursor  *string           `json:"nextCursor"`
	Complete    *bool             `json:"complete"`
	Groups      []json.RawMessage `json:"groups"`
	Cards       []json.RawMessage `json:"cards"`
}
