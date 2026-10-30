package userapi

import (
	"context"
	"encoding/json"
	"slices"
)

type workMetadata struct {
	Links        []*ResourceLink       `json:"links"`
	Attachments  []*ResourceAttachment `json:"attachments"`
	Repositories []*LinkedRepository   `json:"repositories"`
	GitHub       []*LinkedPullRequest  `json:"github"`
	Figma        json.RawMessage       `json:"figma"`
	Details      *struct {
		ID string `json:"id"`
	} `json:"details"`
}

type metadataWork struct {
	workMetadata
	ID          string             `json:"id"`
	Kind        string             `json:"kind"`
	TeamID      string             `json:"teamId"`
	Reasons     []*ReadScopeReason `json:"reasons"`
	PartsStatus []*ReadPartStatus  `json:"partsStatus"`
	ContentPage *ReadBodyPage      `json:"contentPage"`
}

type workMetadataResult struct {
	Metadata workMetadata
	Parts    []*ReadPartStatus
	Version  *string
}

func (r Reader) readWorkMetadata(ctx context.Context, id, teamID, query string, names []string) (*workMetadataResult, error) {
	if id == "" || teamID == "" {
		return nil, ErrInvalidResponse
	}
	var data struct {
		Scope *struct {
			SchemaVersion string          `json:"schemaVersion"`
			RootID        string          `json:"rootId"`
			TeamID        string          `json:"teamId"`
			Status        string          `json:"status"`
			Works         []*metadataWork `json:"works"`
		} `json:"getComponentReadScope"`
	}
	if err := r.read(ctx, query, map[string]string{"componentId": id, "teamId": teamID}, &data); err != nil {
		return nil, err
	}
	scope := data.Scope
	if scope == nil || scope.SchemaVersion != ReadScopeSchemaVersion || scope.RootID != id || scope.TeamID != teamID || (scope.Status != "available" && scope.Status != "partial") {
		return nil, ErrInvalidResponse
	}
	var root *metadataWork
	seen := map[string]bool{}
	for _, work := range scope.Works {
		if work == nil || work.ID == "" || work.TeamID == "" || !validBodyResourceKind(work.Kind) || work.Kind == "" || work.Kind == "meeting" || seen[work.Kind+":"+work.ID] || !validScopeReasons(work.Reasons) {
			return nil, ErrInvalidResponse
		}
		seen[work.Kind+":"+work.ID] = true
		for _, reason := range work.Reasons {
			if reason.Kind == "root" {
				if root != nil || work.ID != id || work.Kind != "component" || work.TeamID != teamID {
					return nil, ErrInvalidResponse
				}
				root = work
			}
		}
	}
	if root == nil {
		return nil, ErrInvalidResponse
	}
	parts, err := workMetadataStatuses(root.PartsStatus, root.ContentPage != nil, names)
	if err != nil {
		return nil, err
	}
	metadata := root.workMetadata
	var version *string
	if root.ContentPage != nil {
		metadata, err = r.workMetadata(ctx, id, teamID, root.ContentPage)
		if err != nil {
			return nil, err
		}
		version = &root.ContentPage.Version
		for _, part := range parts {
			if part.Status == "deferred" {
				part.Status = "available"
				complete := true
				part.Complete = &complete
			}
		}
	}
	return &workMetadataResult{Metadata: metadata, Parts: parts, Version: version}, nil
}

func workMetadataStatuses(parts []*ReadPartStatus, deferred bool, names []string) ([]*ReadPartStatus, error) {
	selected := map[string]*ReadPartStatus{}
	for _, part := range parts {
		if part == nil || part.Complete == nil {
			return nil, ErrInvalidResponse
		}
		if !slices.Contains(names, part.Name) {
			continue
		}
		if selected[part.Name] != nil {
			return nil, ErrInvalidResponse
		}
		switch part.Status {
		case "available", "empty":
			if !*part.Complete {
				return nil, ErrInvalidResponse
			}
		case "unavailable", "denied", "unsupported":
			if *part.Complete {
				return nil, ErrInvalidResponse
			}
		case "deferred":
			if !deferred || *part.Complete {
				return nil, ErrInvalidResponse
			}
		default:
			return nil, ErrInvalidResponse
		}
		copyPart := *part
		selected[part.Name] = &copyPart
	}
	result := make([]*ReadPartStatus, 0, len(names))
	for _, name := range names {
		if selected[name] == nil {
			return nil, ErrInvalidResponse
		}
		result = append(result, selected[name])
	}
	return result, nil
}
