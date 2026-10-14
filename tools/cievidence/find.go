package main

import "fmt"

func findWorkflow(m manifest, workflow, id string) (workflowSpec, error) {
	for _, spec := range m.Workflows {
		if spec.Workflow != workflow {
			continue
		}
		if id != "" && id != spec.ID {
			return workflowSpec{}, fmt.Errorf("workflow %s has id %s, not %s", workflow, spec.ID, id)
		}
		return spec, nil
	}
	return workflowSpec{}, fmt.Errorf("workflow %s not registered in %s", workflow, m.ID)
}
