package userapi

import (
	"context"
	"time"
)

type AgentReadClient interface {
	ReadTaskAgents(context.Context, string, string, string, string, func([]byte, int) error) error
}

type AgentReader struct{ Client AgentReadClient }

type AgentRuntime struct {
	Kind       string    `json:"kind,omitempty"`
	Connection string    `json:"connection"`
	LastSeenAt time.Time `json:"last_seen_at,omitzero"`
}

type PublicAgent struct {
	AgentID            string       `json:"agent_id"`
	Name               string       `json:"name"`
	Description        string       `json:"description,omitempty"`
	RoleStatus         string       `json:"role_status"`
	WorkStatus         string       `json:"work_status"`
	Runtime            AgentRuntime `json:"runtime"`
	Callable           bool         `json:"callable"`
	UnavailableReasons []string     `json:"unavailable_reasons,omitempty"`
	IsCurrentAgent     bool         `json:"is_current_agent,omitempty"`
	UpdatedAt          time.Time    `json:"updated_at,omitzero"`
}

type AgentPage struct {
	Agents        []PublicAgent `json:"agents"`
	Total         int           `json:"total"`
	Cursor        string        `json:"cursor"`
	NextCursor    string        `json:"next_cursor,omitempty"`
	FetchedAt     time.Time     `json:"fetched_at"`
	PolicyChecked bool          `json:"policy_checked"`
}

type AgentDetail struct {
	Agent         PublicAgent `json:"agent"`
	FetchedAt     time.Time   `json:"fetched_at"`
	PolicyChecked bool        `json:"policy_checked"`
}
