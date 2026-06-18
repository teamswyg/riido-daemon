package main

import (
	"errors"
	"fmt"
)

const nativeConfigPlanKind = "native-config-plan"

func run(kind, specPath, templatePath, outPath string) error {
	if err := validateRequest(kind, specPath, templatePath, outPath); err != nil {
		return err
	}
	return generateNativeConfigPlan(specPath, templatePath, outPath)
}

func validateRequest(kind, specPath, templatePath, outPath string) error {
	if kind != nativeConfigPlanKind {
		return fmt.Errorf("riidogen: unsupported kind %q", kind)
	}
	if specPath == "" || templatePath == "" || outPath == "" {
		return errors.New("riidogen: -spec, -template, and -out are required")
	}
	return nil
}
