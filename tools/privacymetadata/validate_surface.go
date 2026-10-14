package main

import "fmt"

const (
	serverFacingSurfaceID   = "c11-server-facing-client-metadata"
	providerStatusSurfaceID = "c10-provider-status-sync-request"
)

func checkSurfacePresent(
	policy PolicySnapshot,
	id string,
	problems []problem,
	checks []ShapeCheck,
) ([]problem, []ShapeCheck) {
	_, ok := findSurface(policy, id)
	checks = append(checks, ShapeCheck{Name: "surface:" + id, OK: ok})
	if !ok {
		problems = append(problems, problem{Message: fmt.Sprintf("missing surface %s", id)})
	}
	return problems, checks
}
