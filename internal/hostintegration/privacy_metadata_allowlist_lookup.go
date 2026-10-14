package hostintegration

import "slices"

func (a PrivacyMetadataAllowlist) Surface(id string) (PrivacyMetadataSurfacePolicy, bool) {
	for _, surface := range a.Surfaces {
		if surface.ID == id {
			return surface, true
		}
	}
	return PrivacyMetadataSurfacePolicy{}, false
}

func (s PrivacyMetadataSurfacePolicy) Allows(path string) bool {
	return slices.Contains(s.AllowedJSONPaths, path)
}

func (s PrivacyMetadataSurfacePolicy) Forbids(path string) bool {
	return slices.Contains(s.ForbiddenJSONPaths, path)
}
