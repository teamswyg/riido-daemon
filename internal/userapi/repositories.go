package userapi

import "context"

const repositoriesQuery = `query RiidoCLIRepositories($componentId:String!,$teamId:String!){getComponentReadScope(componentId:$componentId,teamId:$teamId){schemaVersion rootId teamId status works{id kind teamId reasons{kind relation source} repositories{id fullName isPrivate repositoryUrl source} github{id title number state htmlUrl updatedAt commentsCount aiShortSummary} partsStatus{name status complete} contentPage{` + readBodyFields + `}}}}`

type LinkedRepository struct {
	ID            string  `json:"id"`
	FullName      *string `json:"fullName"`
	IsPrivate     *bool   `json:"isPrivate"`
	RepositoryURL *string `json:"repositoryUrl"`
	Source        *string `json:"source"`
}

type LinkedPullRequest struct {
	ID             string  `json:"id"`
	Title          *string `json:"title"`
	Number         *int    `json:"number"`
	State          *string `json:"state"`
	HTMLURL        *string `json:"htmlUrl"`
	UpdatedAt      *string `json:"updatedAt"`
	CommentsCount  *int    `json:"commentsCount"`
	AIShortSummary *string `json:"aiShortSummary"`
}

type RepositoryList struct {
	WorkID          string               `json:"workId"`
	TeamID          string               `json:"teamId"`
	Repositories    []*LinkedRepository  `json:"repositories"`
	PullRequests    []*LinkedPullRequest `json:"pullRequests"`
	PartsStatus     []*ReadPartStatus    `json:"partsStatus"`
	Complete        bool                 `json:"complete"`
	MetadataVersion *string              `json:"metadataVersion"`
}

func (r Reader) Repositories(ctx context.Context, id, teamID string) (*RepositoryList, error) {
	m, err := r.readWorkMetadata(ctx, id, teamID, repositoriesQuery, []string{"repositories", "github"})
	if err != nil {
		return nil, err
	}
	if m.Metadata.Repositories == nil || m.Metadata.GitHub == nil {
		return nil, ErrInvalidResponse
	}
	result := &RepositoryList{WorkID: id, TeamID: teamID, Repositories: m.Metadata.Repositories, PullRequests: m.Metadata.GitHub, PartsStatus: m.Parts, Complete: true, MetadataVersion: m.Version}
	seen := map[string]bool{}
	for _, repo := range result.Repositories {
		if repo == nil || repo.ID == "" || repo.FullName == nil || repo.IsPrivate == nil || repo.RepositoryURL == nil || repo.Source == nil || seen["repo:"+repo.ID] {
			return nil, ErrInvalidResponse
		}
		seen["repo:"+repo.ID] = true
	}
	for _, pr := range result.PullRequests {
		if pr == nil || pr.ID == "" || pr.Title == nil || pr.Number == nil || pr.State == nil || pr.HTMLURL == nil || pr.CommentsCount == nil || pr.AIShortSummary == nil || seen["pr:"+pr.ID] {
			return nil, ErrInvalidResponse
		}
		seen["pr:"+pr.ID] = true
	}
	for _, part := range m.Parts {
		count := len(result.Repositories)
		if part.Name == "github" {
			count = len(result.PullRequests)
		}
		if (!*part.Complete || part.Status == "empty") && count != 0 {
			return nil, ErrInvalidResponse
		}
		result.Complete = result.Complete && *part.Complete
	}
	return result, nil
}
