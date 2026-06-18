package taskdbplane

import (
	"testing"
	"time"
)

func newTestPlane(t *testing.T, path string) *Plane {
	t.Helper()
	plane, err := New(path)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	now := time.Date(2026, 5, 25, 1, 2, 3, 0, time.UTC)
	plane.now = func() time.Time {
		now = now.Add(time.Second)
		return now
	}
	return plane
}
