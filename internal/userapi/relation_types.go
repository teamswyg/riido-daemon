package userapi

type RelationCategory string

const (
	RelationRelatedTo    RelationCategory = "relatedTo"
	RelationBlocking     RelationCategory = "blocking"
	RelationBlockedBy    RelationCategory = "blockedBy"
	RelationHasDuplicate RelationCategory = "hasDuplicate"
	RelationDuplicatedOf RelationCategory = "duplicatedOf"
)

type RelatedIssueType string

const (
	RelatedIssueComponent RelatedIssueType = "component"
	RelatedIssueBacklog   RelatedIssueType = "backlog"
	RelatedIssueIntake    RelatedIssueType = "intake"
)

// IssueID identifies the relation endpoint; nested resource IDs remain distinct.
type WorkRelation struct {
	RelationCategory RelationCategory `json:"relationCategory"`
	IssueID          string           `json:"issueId"`
	IssueType        RelatedIssueType `json:"issueType"`
	Issue            *RelatedIssue    `json:"issue"`
}

type RelatedIssue struct {
	ID        string                 `json:"id"`
	Component *RelatedComponent      `json:"component"`
	Backlog   *RelatedTitledResource `json:"backlog"`
	Intake    *RelatedTitledResource `json:"intake"`
}

// Required scalar pointers distinguish missing/null values from empty/false.
type RelatedComponent struct {
	ID            string         `json:"id"`
	ComponentType *string        `json:"componentType"`
	Title         *string        `json:"title"`
	IsDone        *bool          `json:"isDone"`
	Status        *RelatedStatus `json:"status"`
}

type RelatedTitledResource struct {
	ID    string  `json:"id"`
	Title *string `json:"title"`
}

type RelatedStatus struct {
	ID           string                `json:"id"`
	StatusType   *string               `json:"statusType"`
	StatusColor  *string               `json:"statusColor"`
	ColorPalette *RelatedStatusPalette `json:"colorPalette"`
}

type RelatedStatusPalette struct {
	ID                  string  `json:"id"`
	ColorHex            *string `json:"colorHex"`
	ColorHexLight       *string `json:"colorHexLight"`
	TextColorHex        *string `json:"textColorHex"`
	TextColorHexLight   *string `json:"textColorHexLight"`
	StatusColorHex      *string `json:"statusColorHex"`
	StatusColorHexLight *string `json:"statusColorHexLight"`
}
