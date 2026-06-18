package taskdbplane

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"time"
)

func loadRuntimeLeaseRegistryOrEmpty(path string) (RuntimeLeaseRegistry, error) {
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return RuntimeLeaseRegistry{
			SchemaVersion: RuntimeLeaseRegistrySchemaVersion,
			Leases:        []RuntimeLeaseRecord{},
		}, nil
	}
	if err != nil {
		return RuntimeLeaseRegistry{}, planeWrapf(ErrTaskDBPlaneRegistry, "lease-registry.load", err, "read runtime lease registry")
	}
	var registry RuntimeLeaseRegistry
	if err := json.Unmarshal(body, &registry); err != nil {
		return RuntimeLeaseRegistry{}, planeWrapf(ErrTaskDBPlaneRegistry, "lease-registry.load", err, "decode runtime lease registry")
	}
	if registry.SchemaVersion != RuntimeLeaseRegistrySchemaVersion {
		return RuntimeLeaseRegistry{}, planeErrorf(ErrTaskDBPlaneRegistry, "lease-registry.load", "runtime lease registry schema mismatch: got %q want %q", registry.SchemaVersion, RuntimeLeaseRegistrySchemaVersion)
	}
	if registry.Leases == nil {
		registry.Leases = []RuntimeLeaseRecord{}
	}
	return registry, nil
}

func saveRuntimeLeaseRegistry(path, taskDBPath string, registry RuntimeLeaseRegistry, now time.Time) error {
	registry.SchemaVersion = RuntimeLeaseRegistrySchemaVersion
	registry.TaskDBPath = taskDBPath
	registry.UpdatedAt = now.UTC()
	sort.Slice(registry.Leases, func(i, j int) bool {
		if registry.Leases[i].TaskID != registry.Leases[j].TaskID {
			return registry.Leases[i].TaskID < registry.Leases[j].TaskID
		}
		return registry.Leases[i].FencingToken < registry.Leases[j].FencingToken
	})
	return writeJSONAtomic(path, registry)
}
