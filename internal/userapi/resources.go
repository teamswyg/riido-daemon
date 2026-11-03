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

func (r Reader) Resources(ctx context.Context, id, teamID string) (*Resources, error) {
	m, err := r.readWorkMetadata(ctx, id, teamID, resourcesQuery, []string{"links", "attachments", "figma"})
	if err != nil {
		return nil, err
	}
	result, err := projectResources(m.Metadata, m.Parts)
	if err != nil {
		return nil, err
	}
	result.WorkID, result.TeamID, result.MetadataVersion = id, teamID, m.Version
	return result, nil
}

func projectResources(metadata workMetadata, parts []*ReadPartStatus) (*Resources, error) {
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
