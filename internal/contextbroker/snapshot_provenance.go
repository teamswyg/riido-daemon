package contextbroker

import (
	"encoding/json"
	"errors"
	"time"
)

// Include metadata in budget accounting. The budget shrinker only shortens
// free text, so snapshot identity and capture time remain intact.
func snapshotProvenance(data json.RawMessage, result Result, query Query) (json.RawMessage, error) {
	if result.SnapshotID == "" {
		return data, nil
	}
	if query.View != viewBase || result.Live || result.CapturedAt.IsZero() || result.SnapshotID != query.AssignmentID {
		return nil, errors.New("invalid snapshot provenance")
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil || fields == nil {
		return nil, errors.New("snapshot context must be an object")
	}
	id, err := json.Marshal(result.SnapshotID)
	if err != nil {
		return nil, err
	}
	capturedAt, err := json.Marshal(result.CapturedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	fields["snapshot_id"], fields["captured_at"] = id, capturedAt
	return json.Marshal(fields)
}
