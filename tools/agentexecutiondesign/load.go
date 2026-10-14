package main

import "path/filepath"

func loadModel(repo, manifestPath string) (model, error) {
	var out model
	if err := readJSON(repoPath(repo, manifestPath), &out.Manifest); err != nil {
		return out, err
	}
	base := manifestBase(manifestPath)
	if err := loadFragments(repo, base, &out); err != nil {
		return out, err
	}
	if err := loadEvidence(repo, &out); err != nil {
		return out, err
	}
	return out, nil
}

func loadFragment(repo, base, rel string, value any) error {
	return readJSON(repoPath(repo, filepath.Join(base, rel)), value)
}

func loadFragments(repo, base string, out *model) error {
	refs := out.Manifest.Fragments
	if err := loadFragment(repo, base, refs.Overview, &out.Overview); err != nil {
		return err
	}
	if err := loadFragment(repo, base, refs.RiskModel, &out.Risk); err != nil {
		return err
	}
	if err := loadFragment(repo, base, refs.ExecutionModel, &out.Execution); err != nil {
		return err
	}
	if err := loadFragment(repo, base, refs.LifecycleModel, &out.Lifecycle); err != nil {
		return err
	}
	return loadFragment(repo, base, refs.Governance, &out.Governance)
}
