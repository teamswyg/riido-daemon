package taskdbplane

import (
	"encoding/json"
	"errors"
	"os"
	"sort"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func (p *Plane) saveRuntimeRegistry() error {
	registry := RuntimeRegistry{
		SchemaVersion: RuntimeRegistrySchemaVersion,
		TaskDBPath:    p.path,
		UpdatedAt:     p.now().UTC(),
		Runtimes:      sortedRuntimeRegistry(p.runtimes),
	}
	return writeJSONAtomic(p.registryPath, registry)
}

func (p *Plane) reloadRuntimeRegistry() error {
	runtimes, err := loadRuntimeRegistryOrEmpty(p.registryPath)
	if err != nil {
		return err
	}
	p.runtimes = runtimes
	return nil
}

func sortedRuntimeRegistry(runtimes map[string]controlplane.RegisteredRuntime) []controlplane.RegisteredRuntime {
	ids := make([]string, 0, len(runtimes))
	for id := range runtimes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]controlplane.RegisteredRuntime, 0, len(ids))
	for _, id := range ids {
		out = append(out, runtimes[id])
	}
	return out
}

func loadRuntimeRegistryOrEmpty(path string) (map[string]controlplane.RegisteredRuntime, error) {
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]controlplane.RegisteredRuntime{}, nil
	}
	if err != nil {
		return nil, planeWrapf(ErrTaskDBPlaneRegistry, "registry.load", err, "read runtime registry")
	}
	var registry RuntimeRegistry
	if err := json.Unmarshal(body, &registry); err != nil {
		return nil, planeWrapf(ErrTaskDBPlaneRegistry, "registry.load", err, "decode runtime registry")
	}
	if registry.SchemaVersion != RuntimeRegistrySchemaVersion {
		return nil, planeErrorf(ErrTaskDBPlaneRegistry, "registry.load", "runtime registry schema mismatch: got %q want %q", registry.SchemaVersion, RuntimeRegistrySchemaVersion)
	}
	out := make(map[string]controlplane.RegisteredRuntime, len(registry.Runtimes))
	for _, runtime := range registry.Runtimes {
		if runtime.RuntimeID != "" {
			out[runtime.RuntimeID] = runtime
		}
	}
	return out, nil
}
