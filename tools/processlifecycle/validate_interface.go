package main

import (
	"fmt"
	"slices"
)

func validateInterfaces(repo string, manifest Manifest) ([]problem, []InterfaceCheck) {
	var problems []problem
	checks := make([]InterfaceCheck, 0, len(manifest.Interfaces))
	for _, spec := range manifest.Interfaces {
		actual, err := interfaceMethods(repo, spec)
		required := append([]string(nil), spec.Methods...)
		slices.Sort(required)
		check := InterfaceCheck{Interface: spec.Name, File: spec.File, Required: required, Actual: actual}
		check.Pass = err == nil && containsAll(actual, required)
		if !check.Pass {
			problems = append(problems, problem{fmt.Sprintf("interface drift: %s", spec.Name)})
		}
		checks = append(checks, check)
	}
	return problems, checks
}

func containsAll(actual, required []string) bool {
	for _, method := range required {
		if !slices.Contains(actual, method) {
			return false
		}
	}
	return true
}
