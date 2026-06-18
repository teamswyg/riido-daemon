package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
	"github.com/teamswyg/riido-daemon/internal/project"
)

func printMwsdSnapshot(ctx context.Context, client mwsdbridge.Client) error {
	snapshot, err := client.FetchSnapshot(ctx)
	if err != nil {
		return err
	}
	return printJSON(snapshot)
}

func printMwsdProjection(ctx context.Context, client mwsdbridge.Client) error {
	projection, err := fetchMwsdProjection(ctx, client)
	if err != nil {
		return err
	}
	return printJSON(projection)
}

func fetchMwsdProjection(ctx context.Context, client mwsdbridge.Client) (project.WorkspaceProjection, error) {
	snapshot, err := client.FetchSnapshot(ctx)
	if err != nil {
		return project.WorkspaceProjection{}, err
	}
	return project.FromMwsdSnapshot(snapshot)
}

func printMwsdMethod[T any](ctx context.Context, client mwsdbridge.Client, method mwsdbridge.Method) error {
	var value T
	if err := client.Request(ctx, string(method), &value); err != nil {
		return err
	}
	return printJSON(value)
}
