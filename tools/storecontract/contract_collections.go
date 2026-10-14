package main

import "strings"

func validateContractCollections(loaded contract) []string {
	var problems []string
	if len(loaded.StoreArtifactRoots) == 0 {
		problems = append(problems, "store_artifact_roots must not be empty")
	}
	problems = append(problems, validateNoticeTermShape(loaded.RequiredNoticeTerms)...)
	return problems
}

func validateNoticeTermShape(terms []string) []string {
	if len(terms) == 0 {
		return []string{"required_notice_terms must not be empty"}
	}
	var problems []string
	for _, term := range terms {
		if strings.TrimSpace(term) == "" {
			problems = append(problems, "required_notice_terms must not include empty terms")
		}
	}
	return problems
}
