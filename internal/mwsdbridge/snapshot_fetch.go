package mwsdbridge

import "context"

// FetchSnapshot reads every mwsd contract Riido needs for its first workspace
// state projection.
func (c Client) FetchSnapshot(ctx context.Context) (Snapshot, error) {
	var snapshot Snapshot
	if err := c.Request(ctx, string(MethodStatus), &snapshot.Status); err != nil {
		return snapshot, err
	}
	if err := c.Request(ctx, string(MethodGraph), &snapshot.Graph); err != nil {
		return snapshot, err
	}
	if err := c.Request(ctx, string(MethodDomain), &snapshot.Domain); err != nil {
		return snapshot, err
	}
	if err := c.Request(ctx, string(MethodHarness), &snapshot.Harness); err != nil {
		return snapshot, err
	}
	if err := c.Request(ctx, string(MethodOrchestration), &snapshot.Orchestration); err != nil {
		return snapshot, err
	}
	if err := c.Request(ctx, string(MethodProjects), &snapshot.Projects); err != nil {
		return snapshot, err
	}
	return snapshot, snapshot.Validate()
}
