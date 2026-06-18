package hostintegration

import "fmt"

func validatePrivacySurface(
	index int,
	surface PrivacyMetadataSurfacePolicy,
	seen map[string]struct{},
) []error {
	var errs []error
	if surface.ID == "" {
		errs = append(errs, fmt.Errorf("surfaces[%d].id is required", index))
	}
	if _, ok := seen[surface.ID]; ok {
		errs = append(errs, fmt.Errorf("surfaces[%d].id duplicates %s", index, surface.ID))
	}
	seen[surface.ID] = struct{}{}
	if surface.OwnerContext == "" {
		errs = append(errs, fmt.Errorf("surfaces[%d].owner_context is required", index))
	}
	if len(surface.AllowedJSONPaths) == 0 {
		errs = append(errs, fmt.Errorf("surfaces[%d].allowed_json_paths is required", index))
	}
	forbidden := forbiddenPrivacyPaths(surface.ForbiddenJSONPaths, index, &errs)
	errs = append(errs, validateAllowedPrivacyPaths(surface.AllowedJSONPaths, forbidden, index)...)
	return errs
}

func forbiddenPrivacyPaths(paths []string, index int, errs *[]error) map[string]struct{} {
	forbidden := map[string]struct{}{}
	for _, path := range paths {
		if path == "" {
			*errs = append(*errs, fmt.Errorf("surfaces[%d].forbidden_json_paths contains empty path", index))
			continue
		}
		forbidden[path] = struct{}{}
	}
	return forbidden
}

func validateAllowedPrivacyPaths(
	paths []string,
	forbidden map[string]struct{},
	index int,
) []error {
	var errs []error
	for _, path := range paths {
		if path == "" {
			errs = append(errs, fmt.Errorf("surfaces[%d].allowed_json_paths contains empty path", index))
			continue
		}
		if _, ok := forbidden[path]; ok {
			errs = append(errs, fmt.Errorf("surfaces[%d] allows forbidden path %s", index, path))
		}
	}
	return errs
}
