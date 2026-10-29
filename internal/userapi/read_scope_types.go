package userapi

import "encoding/json"

const ReadScopeSchemaVersion = "riido-read-scope.v1"

// Product detail objects are selected by a fixed typed GraphQL query. Preserve
// their nullable fields exactly instead of mapping them to CLI display defaults.
type ReadScope struct {
	SchemaVersion string                  `json:"schemaVersion"`
	RootID        string                  `json:"rootId"`
	TeamID        string                  `json:"teamId"`
	Status        string                  `json:"status"`
	Selection     *ReadScopeSelection     `json:"selection"`
	Works         []*ReadScopeWork        `json:"works"`
	Meetings      []*ReadMeetingReference `json:"meetings"`
}

type ReadScopeSelection struct {
	RelatedWorkLimit    int `json:"relatedWorkLimit"`
	RelatedMeetingLimit int `json:"relatedMeetingLimit"`
	ChildrenLimit       int `json:"childrenLimit"`
}

type ReadScopeReason struct {
	Kind     string  `json:"kind"`
	Relation *string `json:"relation"`
	Source   *string `json:"source,omitempty"`
}

type ReadScopeWork struct {
	Sprints         []json.RawMessage     `json:"sprints"`
	Origin          json.RawMessage       `json:"origin"`
	Summary         *ReadResourceSummary  `json:"summary,omitempty"`
	ContentPage     *ReadBodyPage         `json:"contentPage,omitempty"`
	ID              string                `json:"id"`
	Kind            string                `json:"kind"`
	TeamID          string                `json:"teamId"`
	Reasons         []*ReadScopeReason    `json:"reasons"`
	Details         json.RawMessage       `json:"details"`
	BacklogDetails  json.RawMessage       `json:"backlogDetails"`
	IntakeDetails   json.RawMessage       `json:"intakeDetails"`
	Body            *ReadBodyPage         `json:"body"`
	DescriptionBody *ReadBodyPage         `json:"descriptionBody"`
	AlternateBody   *ReadBodyPage         `json:"alternateBody"`
	Comments        *ReadScopeCommentPage `json:"comments"`
	Links           []json.RawMessage     `json:"links"`
	GitHub          []json.RawMessage     `json:"github"`
	Repositories    []json.RawMessage     `json:"repositories"`
	Priorities      []json.RawMessage     `json:"priorities"`
	Checklists      []json.RawMessage     `json:"checklists"`
	Attachments     []json.RawMessage     `json:"attachments"`
	Figma           json.RawMessage       `json:"figma"`
	PartsStatus     []*ReadPartStatus     `json:"partsStatus"`
}

type ReadMeetingReference struct {
	ID       string             `json:"id"`
	TeamID   string             `json:"teamId"`
	Reasons  []*ReadScopeReason `json:"reasons"`
	Snapshot *ReadMeeting       `json:"snapshot,omitempty"`
}

type ReadScopeCommentPage struct {
	ContentPage     *ReadBodyPage     `json:"contentPage,omitempty"`
	ResourceID      string            `json:"resourceId"`
	ResourceKind    string            `json:"resourceKind"`
	ParentCommentID *string           `json:"parentCommentId"`
	Limit           int               `json:"limit"`
	Version         string            `json:"version"`
	Status          string            `json:"status"`
	Page            int               `json:"page"`
	NextPage        *int              `json:"nextPage"`
	Complete        *bool             `json:"complete"`
	Items           []json.RawMessage `json:"items"`
}

type ReadResourceSummary struct {
	BranchName    *string `json:"branchName"`
	ID            string  `json:"id"`
	Kind          string  `json:"kind"`
	Title         string  `json:"title"`
	UpdatedAt     string  `json:"updatedAt"`
	KeyNumber     *string `json:"keyNumber"`
	ComponentType *string `json:"componentType"`
}

type ReadBodyPage struct {
	ResourceID string            `json:"resourceId"`
	Part       string            `json:"part"`
	Status     string            `json:"status"`
	Source     string            `json:"source"`
	Format     string            `json:"format"`
	Version    string            `json:"version"`
	Offset     *int              `json:"offset"`
	Content    string            `json:"content"`
	NextOffset *int              `json:"nextOffset"`
	Complete   *bool             `json:"complete"`
	Selector   *ReadPartSelector `json:"selector,omitempty"`
}
