package main

import (
	"reflect"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/jsontest"
)

func validateShapes(policy PolicySnapshot) ([]problem, []ShapeCheck) {
	var problems []problem
	var checks []ShapeCheck
	problems, checks = checkSurfacePresent(policy, serverFacingSurfaceID, problems, checks)
	problems, checks = checkSurfacePresent(policy, providerStatusSurfaceID, problems, checks)
	problems, checks = checkServerFacingShape(policy, problems, checks)
	return problems, checks
}

func checkServerFacingShape(policy PolicySnapshot, problems []problem, checks []ShapeCheck) ([]problem, []ShapeCheck) {
	surface, _ := findSurface(policy, serverFacingSurfaceID)
	got := jsontest.StructJSONPaths(reflect.TypeFor[hostintegration.ServerFacingClientMetadata]())
	ok := reflect.DeepEqual(got, surface.AllowedJSONPaths)
	checks = append(checks, ShapeCheck{Name: "server-facing-struct-paths", OK: ok})
	if !ok {
		problems = append(problems, problem{Message: "server-facing struct paths drifted from allowlist"})
	}
	return problems, checks
}
