package supervisor

import "testing"

func expectDetachedReportCount(t *testing.T, reports []detachedReport, want int) {
	t.Helper()
	if len(reports) != want {
		t.Fatalf("detached report count = %d, want %d", len(reports), want)
	}
}
