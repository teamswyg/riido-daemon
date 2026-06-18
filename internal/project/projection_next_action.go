package project

type NextAction struct {
	Direction             string `json:"direction"`
	CommandSurface        string `json:"command_surface"`
	Reason                string `json:"reason"`
	RequiresHumanApproval bool   `json:"requires_human_approval"`
}
