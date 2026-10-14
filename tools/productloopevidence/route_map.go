package main

import "fmt"

const routeMapSchema = "riido-entrypoint-route-map.v1"

func loadEntrypointRouteMap(root string, m manifest) (entrypointRouteMap, error) {
	var routes entrypointRouteMap
	if err := loadJSON(repoPath(root, m.EntrypointRouteMap), &routes); err != nil {
		return entrypointRouteMap{}, err
	}
	return routes, validateEntrypointRouteMap(routes)
}

func validateEntrypointRouteMap(routes entrypointRouteMap) error {
	if routes.SchemaVersion != routeMapSchema {
		return fmt.Errorf("entrypoint route map schema_version = %q, want %q", routes.SchemaVersion, routeMapSchema)
	}
	if routes.ID == "" || routes.Title == "" || routes.GeneratedDoc == "" {
		return fmt.Errorf("entrypoint route map id, title, and generated_doc are required")
	}
	if len(routes.Routes) == 0 {
		return fmt.Errorf("entrypoint route map routes are required")
	}
	if err := validateLoop(routes.Loop); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, route := range routes.Routes {
		if route.ID == "" || route.Owner == "" || len(route.Includes) == 0 {
			return fmt.Errorf("entrypoint route map route id, owner, and includes are required")
		}
		if seen[route.ID] {
			return fmt.Errorf("duplicate entrypoint route id %q", route.ID)
		}
		seen[route.ID] = true
	}
	return nil
}
