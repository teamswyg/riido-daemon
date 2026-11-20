package saasplane

import (
	"encoding/hex"
	"errors"
	"strings"
)

// ValidateExecutionCoverage rejects incomplete or mixed enrollment metadata.
// The opaque generation must be copied from the server; it is never normalized.
func ValidateExecutionCoverage(deviceID, deviceSecret, generation string) error {
	covered := strings.HasPrefix(deviceID, "dcov_") || bootstrappedCoverageIdentity(deviceID, generation)
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

// A bootstrap keeps the server's stable native ID. This is wire validation,
// never local authority to promote a legacy registration: the server compares
// the native secret and its committed generation on every covered request.
func bootstrappedCoverageIdentity(deviceID, generation string) bool {
	if !strings.HasPrefix(deviceID, "dev_") || len(deviceID) != 36 || !strings.HasPrefix(generation, "covb_") || len(generation) != 69 {
		return false
	}
	device, deviceErr := hex.DecodeString(deviceID[4:])
	coverage, coverageErr := hex.DecodeString(generation[5:])
	return deviceErr == nil && coverageErr == nil && len(device) == 16 && len(coverage) == 32 && strings.ToLower(deviceID) == deviceID && strings.ToLower(generation) == generation
}
