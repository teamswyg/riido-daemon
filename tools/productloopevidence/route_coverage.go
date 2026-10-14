package main

import "path"

func buildRouteCoverage(root string, routes entrypointRouteMap) routeCoverage {
	entrypoints := listEntrypoints(root)
	rows := make([]routeCoverageRow, 0, len(routes.Routes))
	covered := map[string]bool{}
	for _, route := range routes.Routes {
		count := 0
		for _, entrypoint := range entrypoints {
			if routeMatches(route, entrypoint) {
				covered[entrypoint] = true
				count++
			}
		}
		rows = append(rows, routeCoverageRow{
			ID:              route.ID,
			Owner:           route.Owner,
			EntrypointCount: count,
		})
	}
	uncovered := make([]string, 0)
	for _, entrypoint := range entrypoints {
		if !covered[entrypoint] {
			uncovered = append(uncovered, entrypoint)
		}
	}
	return routeCoverage{
		RouteCount:               len(routes.Routes),
		EntrypointCount:          len(entrypoints),
		CoveredEntrypointCount:   len(entrypoints) - len(uncovered),
		UncoveredEntrypointCount: len(uncovered),
		CoverageRatio:            ratio(len(entrypoints)-len(uncovered), len(entrypoints)),
		UncoveredEntrypoints:     uncovered,
		Routes:                   rows,
	}
}

func routeMatches(route entrypointRoute, entrypoint string) bool {
	for _, include := range route.Includes {
		if ok, _ := path.Match(include, entrypoint); ok {
			return true
		}
	}
	return false
}
