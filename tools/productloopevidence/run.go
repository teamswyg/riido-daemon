package main

import (
	"fmt"
	"os"
)

func run(opts options) error {
	root := repoRoot()
	m, err := loadManifest(repoPath(root, opts.Manifest))
	if err != nil {
		return err
	}
	routes, err := loadEntrypointRouteMap(root, m)
	if err != nil {
		return err
	}
	doc := renderDoc(m, routes)
	routeDoc := renderRouteDoc(routes, m.EntrypointRouteMap)
	if opts.WriteDoc {
		if err := writeText(repoPath(root, m.GeneratedDoc), doc); err != nil {
			return err
		}
		if err := writeText(repoPath(root, routes.GeneratedDoc), routeDoc); err != nil {
			return err
		}
	}
	if opts.CheckDoc {
		if err := checkGeneratedDoc(root, m.GeneratedDoc, doc); err != nil {
			return err
		}
		if err := checkGeneratedDoc(root, routes.GeneratedDoc, routeDoc); err != nil {
			return err
		}
	}
	report, err := buildReport(root, m)
	if err != nil {
		return err
	}
	if opts.EvidenceOut != "" {
		if err := writeJSON(repoPath(root, opts.EvidenceOut), report); err != nil {
			return err
		}
	}
	if report.Status == statusFailed || opts.Strict && report.Status != statusPassed {
		return fmt.Errorf("product loop evidence status=%s problems=%v", report.Status, report.Problems)
	}
	return nil
}

func checkGeneratedDoc(root, path, want string) error {
	current, err := os.ReadFile(repoPath(root, path))
	if err != nil {
		return fmt.Errorf("read generated doc %s: %w", path, err)
	}
	if string(current) != want {
		return fmt.Errorf("generated doc drift: run go run ./tools/productloopevidence -write-doc")
	}
	return nil
}
