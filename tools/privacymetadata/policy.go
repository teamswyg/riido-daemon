package main

import "github.com/teamswyg/riido-daemon/internal/hostintegration"

type privacyPolicyFile struct {
	SchemaVersion string                                         `json:"schema_version"`
	Loop          map[string]string                              `json:"loop,omitempty"`
	Surfaces      []hostintegration.PrivacyMetadataSurfacePolicy `json:"surfaces"`
}

func loadPolicy(repo, rel string) (PolicySnapshot, error) {
	var file privacyPolicyFile
	if err := readJSON(repoPath(repo, rel), &file); err != nil {
		return PolicySnapshot{}, err
	}
	policy := hostintegration.PrivacyMetadataAllowlist{
		SchemaVersion: file.SchemaVersion,
		Surfaces:      file.Surfaces,
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
