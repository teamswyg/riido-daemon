package main

import (
	"fmt"
	"path/filepath"
)

func findWorkflow(m manifest, path string) (workflowSpec, error) {
	path = filepath.ToSlash(path)
	for _, spec := range m.Workflows {
		if filepath.ToSlash(spec.Workflow) == path {
			return spec, nil
		}
	}
	return workflowSpec{}, fmt.Errorf("workflow %s is not registered", path)
}
