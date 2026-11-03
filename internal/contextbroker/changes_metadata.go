package contextbroker

import (
	"encoding/json"
	"errors"
)

// Explicit projection prevents upstream body fields from becoming CLI evidence.
// Pointers distinguish a required false value from an absent contract field.
type changesData struct {
	SnapshotID          string        `json:"snapshot_id"`
	Since               string        `json:"since"`
	Changes             []changedPart `json:"changes"`
	Complete            *bool         `json:"complete"`
	FullRefreshRequired *bool         `json:"full_refresh_required"`
	NextAction          string        `json:"next_action,omitempty"`
}

type changedPart struct {
	ResourceKind string           `json:"resource_kind"`
	SourceID     string           `json:"source_id"`
	Part         string           `json:"part"`
	Status       string           `json:"status"`
	Reason       string           `json:"reason,omitempty"`
	Title        string           `json:"title,omitempty"`
	TeamID       string           `json:"team_id,omitempty"`
	KeyNumber    string           `json:"key_number,omitempty"`
	Reasons      []changeReason   `json:"reasons,omitempty"`
	Current      *changedManifest `json:"current,omitempty"`
}

type changeReason struct {
	Source   *string `json:"source,omitempty"`
	Kind     string  `json:"kind"`
	Relation *string `json:"relation"`
}

type changedManifest struct {
	ResourceKind string   `json:"resource_kind"`
	SourceID     string   `json:"source_id"`
	Part         string   `json:"part"`
	Version      string   `json:"version"`
	Digest       string   `json:"digest,omitempty"`
	Status       string   `json:"status"`
	Complete     *bool    `json:"complete"`
	Cursors      []string `json:"cursors"`
	NextAction   string   `json:"next_action,omitempty"`
}

var errChangesMetadata = errors.New("invalid captured comparison metadata")

func changesMetadata(result Result, query Query) (json.RawMessage, error) {
	var data changesData
	if result.Live || result.CapturedAt.IsZero() || json.Unmarshal(result.Data, &data) != nil ||
		data.SnapshotID != query.AssignmentID || data.Since != query.Since || data.Changes == nil ||
		data.Complete == nil || data.FullRefreshRequired == nil || *data.Complete == *data.FullRefreshRequired ||
		(result.Truncated && *data.Complete) {
		return nil, errChangesMetadata
	}
	seen := map[string]bool{}
	for _, change := range data.Changes {
		key := change.ResourceKind + "\x00" + change.SourceID + "\x00" + change.Part
		if seen[key] || !validChangedPart(change) || (change.Status == "unavailable" && *data.Complete) {
			return nil, errChangesMetadata
		}
		seen[key] = true
	}
	if !validChangeSources(result.Sources, data.Changes) {
		return nil, errChangesMetadata
	}
	data.NextAction = ""
	if *data.FullRefreshRequired {
		data.NextAction = "riido context get"
	}
	return json.Marshal(data)
}
