package main

import (
	"encoding/json"
	"os"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
)

func loadPolicy(repo, rel string) (PolicySnapshot, error) {
	body, err := os.ReadFile(repoPath(repo, rel))
	if err != nil {
		return PolicySnapshot{}, err
	}
	var policy hostintegration.PrivacyMetadataAllowlist
	if err := json.Unmarshal(body, &policy); err != nil {
		return PolicySnapshot{}, err
	}
	if err := policy.Validate(); err != nil {
		return PolicySnapshot{}, err
	}
	return snapshotPolicy(policy), nil
}

func snapshotPolicy(policy hostintegration.PrivacyMetadataAllowlist) PolicySnapshot {
	surfaces := make([]SurfaceSnapshot, len(policy.Surfaces))
	for i, surface := range policy.Surfaces {
		surfaces[i] = SurfaceSnapshot{
			ID:                 surface.ID,
			OwnerContext:       surface.OwnerContext,
			AllowedJSONPaths:   surface.AllowedJSONPaths,
			ForbiddenJSONPaths: surface.ForbiddenJSONPaths,
		}
	}
	return PolicySnapshot{SchemaVersion: policy.SchemaVersion, Surfaces: surfaces}
}
