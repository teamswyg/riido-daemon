package main

import "fmt"

func loadFragments(repo, manifestPath string, m *manifest) error {
	load := func(key string, target any) error {
		rel := m.Fragments[key]
		if rel == "" {
			return fmt.Errorf("missing fragment %q", key)
		}
		return readJSON(fragmentPath(repo, manifestPath, rel), target)
	}
	for key, target := range map[string]any{
		"acl_locations":               &m.ACL,
		"dependency_direction":        &m.Dependency,
		"figma_daemon_boundaries":     &m.FigmaDaemon,
		"figma_onboarding_boundaries": &m.FigmaOnboarding,
		"split_repo_ownership":        &m.SplitRepo,
		"change_procedure":            &m.ChangeProcedure,
	} {
		if err := load(key, target); err != nil {
			return err
		}
	}
	return nil
}
