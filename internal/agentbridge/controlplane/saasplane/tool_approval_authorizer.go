package saasplane

import "context"

func (p *Plane) AuthorizeToolApproval(ctx context.Context, executionID string) (bool, error) {
	_, ok, err := p.assignmentForExecution(ctx, executionID)
	return ok, err
}
