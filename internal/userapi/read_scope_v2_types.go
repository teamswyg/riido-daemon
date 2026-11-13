package userapi

import "encoding/json"

// v2 DTOs are separate so v1 readers and their JSON shape remain unchanged.
type ReadScopeV2 struct {
	ReadScope
	WorkspaceID string             `json:"workspaceId"`
	TeamID      *string            `json:"teamId"`
	Works       []*ReadScopeWorkV2 `json:"works"`
}

type ReadScopeWorkV2 struct {
	ReadScopeWork
	WorkspaceID     string                  `json:"workspaceId"`
	TeamID          *string                 `json:"teamId"`
	ContentPage     *ReadBodyPageV2         `json:"contentPage,omitempty"`
	Body            *ReadBodyPageV2         `json:"body"`
	DescriptionBody *ReadBodyPageV2         `json:"descriptionBody"`
	AlternateBody   *ReadBodyPageV2         `json:"alternateBody"`
	Comments        *ReadScopeCommentPageV2 `json:"comments"`
}

type ReadBodyPageV2 struct {
	ReadBodyPage
	SchemaVersion string  `json:"schemaVersion"`
	WorkspaceID   string  `json:"workspaceId"`
	TeamID        *string `json:"teamId"`
	ResourceKind  string  `json:"resourceKind"`
}

type ReadScopeCommentPageV2 struct {
	ReadScopeCommentPage
	SchemaVersion string          `json:"schemaVersion"`
	WorkspaceID   string          `json:"workspaceId"`
	TeamID        *string         `json:"teamId"`
	ContentPage   *ReadBodyPageV2 `json:"contentPage,omitempty"`
}

type WorkV2 struct {
	Work
	SchemaVersion                string                 `json:"schemaVersion"`
	WorkspaceID                  string                 `json:"workspaceId"`
	OwnerTeamID                  *string                `json:"ownerTeamId"`
	CanEdit                      *bool                  `json:"canEdit"`
	CreatedByWorkspaceMember     *NativeWorkspaceMember `json:"createdByWorkspaceMember"`
	LastUpdatedByWorkspaceMember *NativeWorkspaceMember `json:"lastUpdatedByWorkspaceMember"`
}

func decodeV2Locator(raw []byte, into any, teamField string) error {
	if !nullableFieldPresent(raw, teamField) {
		return ErrInvalidResponse
	}
	return json.Unmarshal(raw, into)
}

func (v *ReadScopeV2) UnmarshalJSON(raw []byte) error {
	type plain ReadScopeV2
	return decodeV2Locator(raw, (*plain)(v), "teamId")
}

func (v *ReadScopeWorkV2) UnmarshalJSON(raw []byte) error {
	type plain ReadScopeWorkV2
	return decodeV2Locator(raw, (*plain)(v), "teamId")
}

func (v *ReadBodyPageV2) UnmarshalJSON(raw []byte) error {
	type plain ReadBodyPageV2
	return decodeV2Locator(raw, (*plain)(v), "teamId")
}

func (v *ReadScopeCommentPageV2) UnmarshalJSON(raw []byte) error {
	type plain ReadScopeCommentPageV2
	return decodeV2Locator(raw, (*plain)(v), "teamId")
}

func (v *WorkV2) UnmarshalJSON(raw []byte) error {
	type plain WorkV2
	return decodeV2Locator(raw, (*plain)(v), "ownerTeamId")
}

// Existing text renderers can consume the common content fields. This view is
// never used to authorize or send a request, and JSON output keeps the v2 DTO.
func (v ReadScopeV2) ContentView() ReadScope {
	result := v.ReadScope
	if v.TeamID != nil {
		result.TeamID = *v.TeamID
	}
	result.Works = make([]*ReadScopeWork, 0, len(v.Works))
	for _, work := range v.Works {
		common := work.common()
		result.Works = append(result.Works, &common)
	}
	return result
}

func (v ReadScopeWorkV2) common() ReadScopeWork {
	result := v.ReadScopeWork
	result.Body, result.DescriptionBody, result.AlternateBody, result.ContentPage = v2BodyCommon(v.Body), v2BodyCommon(v.DescriptionBody), v2BodyCommon(v.AlternateBody), v2BodyCommon(v.ContentPage)
	if v.Comments != nil {
		comments := v.Comments.common()
		result.Comments = &comments
	}
	if v.TeamID != nil {
		result.TeamID = *v.TeamID
	}
	return result
}

func v2BodyCommon(v *ReadBodyPageV2) *ReadBodyPage {
	if v == nil {
		return nil
	}
	return &v.ReadBodyPage
}

func (v ReadScopeCommentPageV2) common() ReadScopeCommentPage {
	result := v.ReadScopeCommentPage
	result.ContentPage = v2BodyCommon(v.ContentPage)
	return result
}

func (v ReadScopeCommentPageV2) ContentView() ReadScopeCommentPage { return v.common() }
