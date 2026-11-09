package saasplane

import (
	"strings"
	"testing"
)

func TestBootstrapCoverageKeepsStableNativeID(t *testing.T) {
	device, generation := "dev_"+strings.Repeat("a", 32), "covb_"+strings.Repeat("b", 64)
	if err := ValidateExecutionCoverage(device, "fixture-native-secret", generation); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, device, secret, generation string }{
		{"ordinary_generation", device, "secret", "cov_" + strings.Repeat("b", 64)},
		{"uppercase", device, "secret", "covb_" + strings.Repeat("B", 64)},
		{"wrong_device", "dev_" + strings.Repeat("a", 16), "secret", generation},
		{"wrong_generation_length", device, "secret", "covb_" + strings.Repeat("b", 63)},
		{"missing_native_secret", device, "", generation},
		{"arbitrary_legacy_id", "legacy", "secret", generation},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateExecutionCoverage(tc.device, tc.secret, tc.generation); err == nil {
				t.Fatal("invalid bootstrap provenance accepted")
			}
		})
	}
}
