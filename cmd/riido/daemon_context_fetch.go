package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
	"github.com/teamswyg/riido-daemon/internal/contextbroker"
)

// saasContextFetcher adapts the SaaS control plane to the broker, keeping
// contextbroker free of control-plane transport details.
type saasContextFetcher struct {
	plane interface {
		FetchAssignmentContext(context.Context, string, saasplane.AssignmentContextRequest) (saasplane.AssignmentContextResponse, error)
	}
}

func (f saasContextFetcher) FetchAssignmentContext(ctx context.Context, q contextbroker.Query) (contextbroker.Result, error) {
	out, err := f.plane.FetchAssignmentContext(ctx, q.AgentID, saasplane.AssignmentContextRequest{
		AssignmentID: q.AssignmentID, LeaseToken: q.LeaseToken, View: q.View, Target: q.Target, Cursor: q.Cursor, RequestID: q.RequestID,
	})
	if err != nil {
		if status, message, ok := saasplane.HTTPStatus(err); ok {
			return contextbroker.Result{}, &contextbroker.FetchError{Status: status, Message: message}
		}
		return contextbroker.Result{}, err
	}
	result := contextbroker.Result{Data: out.Data, CapturedAt: out.CapturedAt, Truncated: out.Truncated}
	for _, source := range out.Sources {
		result.Sources = append(result.Sources, contextbroker.ResultSource{Kind: source.Kind, ID: source.ID, Part: source.Part, Version: source.Version})
	}
	return result, nil
}

// daemonContextFetcher returns a fetcher when the control plane is SaaS.
func daemonContextFetcher(reporter any) contextbroker.Fetcher {
	if plane, ok := reporter.(*saasplane.Plane); ok {
		return saasContextFetcher{plane: plane}
	}
	return nil
}
