package main

import (
	"path/filepath"
)

func loadManifest(repo, rel string) (manifest, error) {
	m, err := loadJSON[manifest](repoPath(repo, rel))
	if err != nil {
		return manifest{}, err
	}
	validation, err := loadProviderValidation(repo, m.ProviderValidationManifest)
	if err != nil {
		return manifest{}, err
	}
	realCLI, err := loadJSON[realCLIObservation](repoPath(repo, m.RealCLIObservationManifest))
	if err != nil {
		return manifest{}, err
	}
	m.ProviderValidation = validation
	m.RealCLIObservation = realCLI
	return m, nil
}

func loadProviderValidation(repo, rel string) (providerValidation, error) {
	rootPath := repoPath(repo, rel)
	loaded, err := loadJSON[providerValidation](rootPath)
	if err != nil {
		return providerValidation{}, err
	}
	for _, file := range loaded.ProviderFiles {
		provider, err := loadJSON[providerEvidence](filepath.Join(filepath.Dir(rootPath), filepath.FromSlash(file)))
		if err != nil {
			return providerValidation{}, err
		}
		loaded.Providers = append(loaded.Providers, provider)
	}
	return loaded, nil
}
