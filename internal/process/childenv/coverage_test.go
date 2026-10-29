package childenv

import "testing"

func TestCoverageMetadataStaysOutOfChildEnvironment(t *testing.T) {
	got := Scrub([]string{"PATH=/bin", "RIIDO_EXECUTION_COVERAGE_GENERATION=cov_fixture", "riido_execution_coverage_generation=cov_other"})
	if len(got) != 1 || got[0] != "PATH=/bin" {
		t.Fatal("private enrollment metadata leaked to provider child")
	}
}
