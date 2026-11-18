package userapi

import (
	"context"
	"strings"
)

var (
	relationsV2Query = strings.Replace(strings.Replace(relationsQuery, "$teamId:String!", "$teamId:String,$workspaceId:String!", 1), "teamId:$teamId)", "teamId:$teamId,workspaceId:$workspaceId,locatorVersion:2)", 1)
	activityV2Query  = strings.Replace(strings.Replace(activityQuery, "$teamId:String!", "$teamId:String,$workspaceId:String!", 1), "teamId:$teamId,", "teamId:$teamId,workspaceId:$workspaceId,locatorVersion:2,", 1)
)

func (r Reader) RelationsV2(ctx context.Context, locator WorkLocator) ([]*WorkRelation, error) {
	if !locator.valid() {
		return nil, ErrInvalidRequest
	}
	var data struct {
		Relations []*WorkRelation `json:"getComponentRelatedIssues"`
	}
	if err := r.read(ctx, relationsV2Query, locator.variables(), &data); err != nil {
		return nil, err
	}
	if !validRelations(data.Relations) {
		return nil, ErrInvalidResponse
	}
	return data.Relations, nil
}

func (r Reader) ActivityV2(ctx context.Context, locator WorkLocator, pagination ActivityPagination) (ActivityPage, error) {
	if !locator.valid() || !pagination.Valid() {
		return ActivityPage{}, ErrInvalidRequest
	}
	var data struct {
		Page *ActivityPage `json:"getComponentActivityLogs"`
	}
	variables := locator.variables()
	variables["pagination"] = pagination
	if err := r.read(ctx, activityV2Query, variables, &data); err != nil {
		return ActivityPage{}, err
	}
	return checkedActivityPage(data.Page, pagination)
}
