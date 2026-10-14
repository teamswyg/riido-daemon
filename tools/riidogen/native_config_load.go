package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func loadNativeConfigPlan(path string) (nativeConfigPlanCatalogSpec, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nativeConfigPlanCatalogSpec{}, fmt.Errorf("riidogen: read native config plan: %w", err)
	}
	var spec nativeConfigPlanCatalogSpec
	if err := json.Unmarshal(body, &spec); err != nil {
		return nativeConfigPlanCatalogSpec{}, fmt.Errorf("riidogen: decode native config plan: %w", err)
	}
	spec.SpecPath = filepath.Base(path)
	if err := validateNativeConfigPlan(spec); err != nil {
		return nativeConfigPlanCatalogSpec{}, err
	}
	sortNativeConfigProviders(spec.Providers)
	return spec, nil
}
