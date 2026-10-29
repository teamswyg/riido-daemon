package userapi

type WorkStatus struct {
	ID         string `json:"id"`
	StatusName string `json:"statusName"`
	StatusType string `json:"statusType"`
}

type WorkMember struct {
	ID          string          `json:"id"`
	MemberOrder int             `json:"memberOrder"`
	TeamMember  *WorkTeamMember `json:"teamMember"`
}

type WorkTeamMember struct {
	ID               string                `json:"id"`
	Status           string                `json:"status"`
	WorkspaceProfile *WorkWorkspaceProfile `json:"workspaceProfile"`
}

type WorkWorkspaceProfile struct {
	Nickname string `json:"nickname"`
	Status   string `json:"status"`
}

type WorkLabel struct {
	ID         string         `json:"id"`
	LabelOrder int            `json:"labelOrder"`
	Label      *WorkLabelInfo `json:"label"`
}

type WorkLabelInfo struct {
	ID        string `json:"id"`
	LabelName string `json:"labelName"`
}
