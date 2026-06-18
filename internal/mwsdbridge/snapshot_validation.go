package mwsdbridge

import (
	"errors"
	"fmt"
)

// Validate checks the schema-level handshake between Riido and mwsd.
func (s Snapshot) Validate() error {
	checks := []struct {
		name string
		got  string
		want string
	}{
		{"graph", s.Graph.SchemaVersion, GraphSchemaVersion},
		{"domain", s.Domain.SchemaVersion, DomainSchemaVersion},
		{"harness", s.Harness.SchemaVersion, HarnessSchemaVersion},
		{"orchestration", s.Orchestration.SchemaVersion, OrchestrationSchemaVersion},
		{"projects", s.Projects.SchemaVersion, ProjectsSchemaVersion},
	}
	for _, check := range checks {
		if check.got != check.want {
			return fmt.Errorf("%s schema mismatch: got %q want %q", check.name, check.got, check.want)
		}
	}
	if s.Status.Root == "" {
		return errors.New("status root is empty")
	}
	if s.Graph.Root != "" && s.Graph.Root != s.Status.Root {
		return fmt.Errorf("graph root mismatch: %s != %s", s.Graph.Root, s.Status.Root)
	}
	if s.Status.OrchestrationSchemaVersion != "" && s.Status.OrchestrationSchemaVersion != OrchestrationSchemaVersion {
		return fmt.Errorf("status orchestration schema mismatch: got %q want %q", s.Status.OrchestrationSchemaVersion, OrchestrationSchemaVersion)
	}
	if s.Orchestration.Root != "" && s.Orchestration.Root != s.Status.Root {
		return fmt.Errorf("orchestration root mismatch: %s != %s", s.Orchestration.Root, s.Status.Root)
	}
	if s.Orchestration.DomainSchemaVersion != "" && s.Orchestration.DomainSchemaVersion != DomainSchemaVersion {
		return fmt.Errorf("orchestration domain schema mismatch: got %q want %q", s.Orchestration.DomainSchemaVersion, DomainSchemaVersion)
	}
	if s.Orchestration.HarnessSchemaVersion != "" && s.Orchestration.HarnessSchemaVersion != HarnessSchemaVersion {
		return fmt.Errorf("orchestration harness schema mismatch: got %q want %q", s.Orchestration.HarnessSchemaVersion, HarnessSchemaVersion)
	}
	if s.Orchestration.TopDownCount != s.Harness.TopDownCount {
		return fmt.Errorf("orchestration top-down count mismatch: %d != %d", s.Orchestration.TopDownCount, s.Harness.TopDownCount)
	}
	if s.Orchestration.BottomUpCount != s.Harness.BottomUpCount {
		return fmt.Errorf("orchestration bottom-up count mismatch: %d != %d", s.Orchestration.BottomUpCount, s.Harness.BottomUpCount)
	}
	if s.Orchestration.NextAction.Direction != "" && s.Orchestration.NextAction.Direction != s.Harness.NextDirection {
		return fmt.Errorf("orchestration next direction mismatch: %s != %s", s.Orchestration.NextAction.Direction, s.Harness.NextDirection)
	}
	if s.Projects.RepositoryCount != len(s.Projects.Repositories) {
		return fmt.Errorf("project registry count mismatch: %d != %d", s.Projects.RepositoryCount, len(s.Projects.Repositories))
	}
	return nil
}
