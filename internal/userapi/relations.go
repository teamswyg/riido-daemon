package userapi

import "context"

const relationsQuery = `query RiidoCLIRelations($componentId:String!,$teamId:String!){getComponentRelatedIssues(componentId:$componentId,teamId:$teamId){relationCategory issueId issueType issue{id component{id componentType title isDone status{id statusType statusColor colorPalette{id colorHex colorHexLight textColorHex textColorHexLight statusColorHex statusColorHexLight}}} backlog{id title} intake{id title}}}}`

// Relations reads direct live links, not a recursive graph or assignment snapshot.
func (r Reader) Relations(ctx context.Context, id, teamID string) ([]*WorkRelation, error) {
	if id == "" || teamID == "" {
		return nil, ErrInvalidResponse
	}
	var data struct {
		Relations []*WorkRelation `json:"getComponentRelatedIssues"`
	}
	if err := r.read(ctx, relationsQuery, map[string]string{"componentId": id, "teamId": teamID}, &data); err != nil {
		return nil, err
	}
	if !validRelations(data.Relations) {
		return nil, ErrInvalidResponse
	}
	return data.Relations, nil
}

func validRelations(rows []*WorkRelation) bool {
	if rows == nil {
		return false
	}
	for _, row := range rows {
		if row == nil || row.Issue == nil || row.IssueID == "" || row.IssueID != row.Issue.ID {
			return false
		}
		switch row.RelationCategory {
		case RelationRelatedTo, RelationBlocking, RelationBlockedBy, RelationHasDuplicate, RelationDuplicatedOf:
		default:
			return false
		}
		issue := row.Issue
		switch row.IssueType {
		case RelatedIssueComponent:
			if issue.Backlog != nil || issue.Intake != nil {
				return false
			}
		case RelatedIssueBacklog:
			if issue.Component != nil || issue.Intake != nil {
				return false
			}
		case RelatedIssueIntake:
			if issue.Component != nil || issue.Backlog != nil {
				return false
			}
		default:
			return false
		}
		if !validRelatedComponent(issue.Component) || !validRelatedTitle(issue.Backlog) || !validRelatedTitle(issue.Intake) {
			return false
		}
	}
	return true
}

func validRelatedTitle(value *RelatedTitledResource) bool {
	return value == nil || (value.ID != "" && value.Title != nil)
}

func validRelatedComponent(value *RelatedComponent) bool {
	if value == nil {
		return true
	}
	if value.ID == "" || value.ComponentType == nil || value.Title == nil || value.IsDone == nil {
		return false
	}
	status := value.Status
	if status == nil {
		return true
	}
	if status.ID == "" || status.StatusType == nil || status.StatusColor == nil {
		return false
	}
	palette := status.ColorPalette
	return palette == nil || (palette.ID != "" && palette.ColorHex != nil && palette.ColorHexLight != nil && palette.TextColorHex != nil && palette.TextColorHexLight != nil && palette.StatusColorHex != nil && palette.StatusColorHexLight != nil)
}
