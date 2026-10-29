package saasplane

import (
	"errors"
	"strings"
)

// ValidateExecutionCoverage rejects incomplete or mixed enrollment metadata.
// The opaque generation must be copied from the server; it is never normalized.
func ValidateExecutionCoverage(deviceID, deviceSecret, generation string) error {
	covered := strings.HasPrefix(deviceID, "dcov_")
	if !covered && generation == "" {
		return nil
	}
	if !covered || deviceSecret == "" || generation == "" || len(generation) > 512 {
		return errors.New("saasplane: covered device credentials require RIIDO_EXECUTION_COVERAGE_GENERATION; legacy devices must omit it")
	}
	for _, c := range generation {
		if c < 0x21 || c > 0x7e || c == '/' || c == '\\' {
			return errors.New("saasplane: invalid RIIDO_EXECUTION_COVERAGE_GENERATION")
		}
	}
	return nil
}
