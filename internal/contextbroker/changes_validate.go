package contextbroker

import (
	"encoding/hex"
	"slices"
	"strings"
)

func validChangedPart(change changedPart) bool {
	if !slices.Contains([]string{"component", "backlog", "intake", "meeting"}, change.ResourceKind) ||
		strings.TrimSpace(change.SourceID) == "" || strings.ContainsRune(change.SourceID, '\x00') ||
		!slices.Contains([]string{"metadata", "editor", "description", "alternate", "comments", "board"}, change.Part) {
		return false
	}
	current := change.Current
	if current == nil {
		// Omitted sources never carry metadata. deleted and unselected are
		// explained by the server; anything else must stay an unknown omission.
		if change.Title != "" || change.TeamID != "" || change.KeyNumber != "" || len(change.Reasons) != 0 {
			return false
		}
		if change.Status == "deleted" || change.Status == "unselected" {
			return change.Reason == ""
		}
		return change.Status == "unavailable" && change.Reason == "not_in_current_capture"
	}
	if current.ResourceKind != change.ResourceKind || current.SourceID != change.SourceID || current.Part != change.Part ||
		current.Complete == nil || !validChangeCursors(current) {
		return false
	}
	if change.Status == "unavailable" {
		return change.Reason == "incomplete_comparison"
	}
	if change.Status != "updated" || change.Reason != "" || !*current.Complete || current.Version == "" ||
		(current.Status != "available" && current.Status != "empty") ||
		!strings.HasPrefix(current.Digest, "sha256:") || len(current.Digest) != 71 {
		return false
	}
	_, err := hex.DecodeString(current.Digest[7:])
	return err == nil
}

func validChangeSources(sources []ResultSource, changes []changedPart) bool {
	allowed := map[string]bool{}
	for _, change := range changes {
		if change.Current != nil {
			allowed[changeSourceKind(change)+"\x00"+change.SourceID] = true
		}
	}
	for _, source := range sources {
		if source.Part != "title" || source.Version != "" || !allowed[source.Kind+"\x00"+source.ID] {
			return false
		}
	}
	return true
}

// Match the server's title audit classification, including hierarchy reasons.
func changeSourceKind(change changedPart) string {
	kind := "related_" + change.ResourceKind
	if change.ResourceKind == "component" {
		kind = "related_task"
	}
	for _, reason := range change.Reasons {
		switch reason.Kind {
		case "root":
			kind = "task"
		case "project":
			kind = "project"
		case "milestone":
			kind = "milestone"
		case "parentTask":
			kind = "parent_task"
		}
	}
	return kind
}

func validChangeCursors(current *changedManifest) bool {
	if current.Cursors == nil || (*current.Complete && len(current.Cursors) == 0) {
		return false
	}
	seen := map[string]bool{}
	for _, cursor := range current.Cursors {
		if strings.TrimSpace(cursor) == "" || seen[cursor] {
			return false
		}
		seen[cursor] = true
	}
	return true
}
