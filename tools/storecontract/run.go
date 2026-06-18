package main

import (
	"errors"
	"fmt"
	"path/filepath"
)

func run(repoRoot, contractPath string) (checkResult, error) {
	result := newCheckResult(contractPath)
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("resolve repo root: %v", err))
		return result, errors.New("invalid repo root")
	}

	loaded, err := loadContract(resolvePath(root, contractPath))
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	result.addContractMetadata(loaded)
	problems := validateStoreContract(root, loaded)
	if len(problems) > 0 {
		result.Errors = problems
		return result, errors.New("store distribution contract failed")
	}

	result.Status = "passed"
	return result, nil
}

func validateStoreContract(repoRoot string, loaded contract) []string {
	var problems []string
	problems = append(problems, validateContractShape(loaded)...)
	problems = append(problems, validateRequiredDocs(repoRoot, loaded.RequiredDocs)...)
	problems = append(problems, validateRequiredNoticeTerms(repoRoot, loaded.RequiredNoticeTerms)...)
	return append(problems, scanArtifactRoots(
		repoRoot,
		loaded.StoreArtifactRoots,
		loaded.ExternalProviderCLINames,
	)...)
}
