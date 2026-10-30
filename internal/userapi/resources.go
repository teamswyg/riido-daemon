package userapi

import (
	"context"
	"encoding/json"
)

const resourcesQuery = `query RiidoCLIResources($componentId:String!,$teamId:String!){getComponentReadScope(componentId:$componentId,teamId:$teamId){schemaVersion rootId teamId status works{id kind teamId reasons{kind relation source} links{id title url} attachments{id fileName fileLocation} figma{id url fileKey nodeId} partsStatus{name status complete} contentPage{` + readBodyFields + `}}}}`

type ResourceLink struct {
	ID    string  `json:"id"`
	Title *string `json:"title"`
	URL   *string `json:"url"`
}

type ResourceAttachment struct {
	ID           string  `json:"id"`
	FileName     *string `json:"fileName"`
	FileLocation *string `json:"fileLocation"`
}

type ResourceFigma struct {
	ID      string  `json:"id"`
	URL     *string `json:"url"`
	FileKey *string `json:"fileKey"`
	NodeID  *string `json:"nodeId"`
}

type Resources struct {
	WorkID          string                `json:"workId"`
	TeamID          string                `json:"teamId"`
	Links           []*ResourceLink       `json:"links"`
	Attachments     []*ResourceAttachment `json:"attachments"`
	Figma           *ResourceFigma        `json:"figma"`
	PartsStatus     []*ReadPartStatus     `json:"partsStatus"`
	Complete        bool                  `json:"complete"`
	MetadataVersion *string               `json:"metadataVersion"`
}

type resourceMetadata struct {
	Links       []*ResourceLink       `json:"links"`
	Attachments []*ResourceAttachment `json:"attachments"`
	Figma       json.RawMessage       `json:"figma"`
	Details     *struct {
		ID string `json:"id"`
	} `json:"details"`
}

type resourceWork struct {
	resourceMetadata
	ID          string             `json:"id"`
	Kind        string             `json:"kind"`
	TeamID      string             `json:"teamId"`
	Reasons     []*ReadScopeReason `json:"reasons"`
	PartsStatus []*ReadPartStatus  `json:"partsStatus"`
	ContentPage *ReadBodyPage      `json:"contentPage"`
}

func (r Reader) Resources(ctx context.Context, id, teamID string) (*Resources, error) {
	if id == "" || teamID == "" {
		return nil, ErrInvalidResponse
	}
	var data struct {
		Scope *struct {
			SchemaVersion string          `json:"schemaVersion"`
			RootID        string          `json:"rootId"`
			TeamID        string          `json:"teamId"`
			Status        string          `json:"status"`
			Works         []*resourceWork `json:"works"`
		} `json:"getComponentReadScope"`
	}
	if err := r.read(ctx, resourcesQuery, map[string]string{"componentId": id, "teamId": teamID}, &data); err != nil {
		return nil, err
	}
	scope := data.Scope
	if scope == nil || scope.SchemaVersion != ReadScopeSchemaVersion || scope.RootID != id || scope.TeamID != teamID || (scope.Status != "available" && scope.Status != "partial") {
		return nil, ErrInvalidResponse
	}
	var root *resourceWork
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
	parts, err := resourceStatuses(root.PartsStatus, root.ContentPage != nil)
	if err != nil {
		return nil, err
	}
	metadata := root.resourceMetadata
	var version *string
	if root.ContentPage != nil {
		metadata, err = r.resourceMetadata(ctx, id, teamID, root.ContentPage)
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
	result, err := projectResources(metadata, parts)
	if err != nil {
		return nil, err
	}
	result.WorkID, result.TeamID, result.MetadataVersion = id, teamID, version
	return result, nil
}

func resourceStatuses(parts []*ReadPartStatus, deferred bool) ([]*ReadPartStatus, error) {
	selected := map[string]*ReadPartStatus{}
	for _, part := range parts {
		if part == nil || part.Complete == nil {
			return nil, ErrInvalidResponse
		}
		if part.Name != "links" && part.Name != "attachments" && part.Name != "figma" {
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
	result := make([]*ReadPartStatus, 0, 3)
	for _, name := range []string{"links", "attachments", "figma"} {
		if selected[name] == nil {
			return nil, ErrInvalidResponse
		}
		result = append(result, selected[name])
	}
	return result, nil
}

func projectResources(metadata resourceMetadata, parts []*ReadPartStatus) (*Resources, error) {
	if metadata.Links == nil || metadata.Attachments == nil || metadata.Figma == nil {
		return nil, ErrInvalidResponse
	}
	result := &Resources{Links: metadata.Links, Attachments: metadata.Attachments, PartsStatus: parts, Complete: true}
	if json.Unmarshal(metadata.Figma, &result.Figma) != nil {
		return nil, ErrInvalidResponse
	}
	if result.Figma != nil && (result.Figma.ID == "" || result.Figma.URL == nil) {
		return nil, ErrInvalidResponse
	}
	seen := map[string]bool{}
	for _, link := range result.Links {
		if link == nil || link.ID == "" || seen["link:"+link.ID] {
			return nil, ErrInvalidResponse
		}
		seen["link:"+link.ID] = true
	}
	for _, file := range result.Attachments {
		if file == nil || file.ID == "" || file.FileName == nil || file.FileLocation == nil || seen["file:"+file.ID] {
			return nil, ErrInvalidResponse
		}
		seen["file:"+file.ID] = true
	}
	for _, part := range parts {
		count := len(result.Links)
		if part.Name == "attachments" {
			count = len(result.Attachments)
		}
		if part.Name == "figma" {
			count = 0
			if result.Figma != nil {
				count = 1
			}
		}
		if (!*part.Complete || part.Status == "empty") && count != 0 {
			return nil, ErrInvalidResponse
		}
		result.Complete = result.Complete && *part.Complete
	}
	return result, nil
}
