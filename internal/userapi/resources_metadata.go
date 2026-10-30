package userapi

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const maxResourceMetadataBytes = 16 << 20

var ErrResourceMetadataTooLarge = errors.New("resource metadata requires manual byte paging")

// ResourceMetadataLimitError retains the already-authorized version so manual
// continuation never depends on another full context (or related meeting) read.
type ResourceMetadataLimitError struct {
	Version string
}

func (e *ResourceMetadataLimitError) Error() string { return ErrResourceMetadataTooLarge.Error() }
func (e *ResourceMetadataLimitError) Unwrap() error { return ErrResourceMetadataTooLarge }

func (r Reader) workMetadata(ctx context.Context, id, team string, first *ReadBodyPage) (workMetadata, error) {
	var metadata workMetadata
	if !validDeferredPart(first, id, "component", "metadata", nil) {
		return metadata, ErrInvalidResponse
	}
	q := ReadBodyRequest{ComponentID: id, TeamID: team, Kind: "component", Part: "metadata", Version: first.Version, Limit: MaxReadBodyLimit}
	page := *first
	var raw strings.Builder
	for {
		if err := ctx.Err(); err != nil {
			return metadata, err
		}
		if page.Source != "riido_metadata" || page.Format != "json" {
			return metadata, ErrInvalidResponse
		}
		if len(page.Content) > maxResourceMetadataBytes-raw.Len() {
			return metadata, &ResourceMetadataLimitError{Version: first.Version}
		}
		raw.WriteString(page.Content)
		if *page.Complete {
			break
		}
		q.Offset = *page.NextOffset
		next, err := r.Body(ctx, q)
		if err != nil {
			return metadata, err
		}
		page = next
	}
	canonical, err := json.Marshal([]string{"riido-read-part.v1", "component", id, team, "metadata", "{}", "riido_metadata", "json", raw.String()})
	if err != nil {
		return metadata, ErrInvalidResponse
	}
	if fmt.Sprintf("sha256:%x", sha256.Sum256(canonical)) != first.Version || json.Unmarshal([]byte(raw.String()), &metadata) != nil || metadata.Details == nil || metadata.Details.ID != id {
		return workMetadata{}, ErrInvalidResponse
	}
	return metadata, nil
}
