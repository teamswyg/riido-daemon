package userapi

import "encoding/json"

func validLocatorVersion(version int) bool { return version == 0 || version == 1 || version == 2 }

func locatorSchema(prefix string, version int) string {
	if version == 2 {
		return prefix + ".v2"
	}
	return prefix + ".v1"
}

func nullableFieldPresent(raw json.RawMessage, field string) bool {
	var fields map[string]json.RawMessage
	return json.Unmarshal(raw, &fields) == nil && len(fields[field]) > 0
}

func validLocatorTeam(version int, kind, componentType string, team *string) bool {
	if version == 2 && kind == "component" {
		if componentType == "project" || componentType == "milestone" {
			return team == nil
		}
		if componentType != "task" && componentType != "subtask" {
			return false
		}
	}
	return team != nil && *team != ""
}

func sameLocatorTeam(left, right *string) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

// WorkLocator is returned only after current server authorization. It is not
// persisted, and supplies no substitute team or authority to a later request.
type WorkLocator struct {
	WorkspaceID   string  `json:"workspaceId"`
	TeamID        *string `json:"teamId"`
	ID            string  `json:"id"`
	ComponentType string  `json:"componentType"`
}

func (w WorkLocator) valid() bool {
	return workspaceIDPattern.MatchString(w.WorkspaceID) && resolvedWorkIDPattern.MatchString(w.ID) &&
		(w.ComponentType == "project" || w.ComponentType == "milestone" || w.ComponentType == "task" || w.ComponentType == "subtask") &&
		validLocatorTeam(2, "component", w.ComponentType, w.TeamID) && (w.TeamID == nil || resolvedWorkIDPattern.MatchString(*w.TeamID))
}

func (w ResolvedWorkReference) Locator() WorkLocator {
	return WorkLocator{WorkspaceID: w.WorkspaceID, TeamID: w.TeamID, ID: w.ID, ComponentType: w.ComponentType}
}

func (w WorkLocator) variables() map[string]any {
	return map[string]any{"componentId": w.ID, "workspaceId": w.WorkspaceID, "teamId": w.TeamID}
}
