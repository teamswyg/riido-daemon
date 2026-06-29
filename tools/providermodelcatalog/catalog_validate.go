package main

import "fmt"

var requiredProviders = []string{"codex", "cursor", "openclaw", "claude"}

func validateCatalog(catalog catalog) error {
	for _, provider := range requiredProviders {
		models := catalog.Providers[provider]
		if len(models) <= 1 {
			return fmt.Errorf(
				"provider model catalog too small: provider=%s count=%d requirement=model_count>1",
				provider,
				len(models),
			)
		}
	}
	return nil
}
