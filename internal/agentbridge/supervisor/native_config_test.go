package supervisor

import "testing"

func TestSupervisorKeepsOpenClawAndCursorNativeConfigInstructionOnly(t *testing.T) {
	for _, provider := range nativeConfigInstructionOnlyProviders() {
		t.Run(string(provider), func(t *testing.T) {
			run := runNativeConfigInstructionOnlyTask(t, provider)

			assertNativeConfigHomeMetadataOmitted(t, run)
			assertNativeConfigInstructionOnlyManifest(t, run)
			assertProviderNativeArtifactsAbsent(t, run)
		})
	}
}
