package userapi

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"
)

func validAgentPage(raw []byte, page AgentPage, cursor string) bool {
	if cursor == "" {
		cursor = "0"
	}
	if !hasThreadFields(raw, "agents", "total", "cursor", "fetched_at", "policy_checked") || page.Agents == nil || len(page.Agents) > 50 || page.Total < 0 || page.Cursor != cursor || page.FetchedAt.IsZero() || page.PolicyChecked {
		return false
	}
	n, _ := strconv.Atoi(cursor) // The caller already validated canonical input.
	// Divide before multiplying/adding so a maximum-sized integer is safe.
	pages := page.Total / 50
	if page.Total%50 != 0 {
		pages++
	}
	want, next := 0, ""
	if n < pages {
		want = min(50, page.Total-n*50)
		if n < pages-1 {
			next = strconv.Itoa(n + 1)
		}
	}
	if len(page.Agents) != want || page.NextCursor != next {
		return false
	}
	var shape struct {
		Agents []json.RawMessage `json:"agents"`
	}
	if json.Unmarshal(raw, &shape) != nil {
		return false
	}
	seen := map[string]bool{}
	for i, agent := range page.Agents {
		if seen[agent.AgentID] || !validPublicAgent(shape.Agents[i], agent) {
			return false
		}
		seen[agent.AgentID] = true
	}
	return true
}

func validPublicAgent(raw []byte, agent PublicAgent) bool {
	if !hasThreadFields(raw, "agent_id", "name", "role_status", "work_status", "runtime", "callable") || !ValidThreadReadID(agent.AgentID) || agent.IsCurrentAgent {
		return false
	}
	var shape struct {
		Runtime json.RawMessage `json:"runtime"`
	}
	if json.Unmarshal(raw, &shape) != nil || !hasThreadFields(shape.Runtime, "connection") {
		return false
	}
	role := "described"
	if strings.TrimSpace(agent.Description) == "" {
		role = "missing"
	}
	if agent.RoleStatus != role || !slices.Contains([]string{"idle", "queued", "running", "waiting_for_user", "completed", "failed", "offline"}, agent.WorkStatus) || !slices.Contains([]string{"", "codex", "claude_code", "cursor", "openclaw"}, agent.Runtime.Kind) {
		return false
	}
	switch agent.Runtime.Connection {
	case "online":
		return agent.Callable && len(agent.UnavailableReasons) == 0
	case "offline":
		return !agent.Callable && slices.Equal(agent.UnavailableReasons, []string{"runtime_offline"})
	default:
		return false
	}
}
