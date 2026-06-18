package controlplane

import "testing"

func TestFileQueueSourceMissingDir(t *testing.T) {
	_, err := NewFileQueueSource("/nonexistent-path-xyz-9999")
	if err == nil {
		t.Fatal("expected error for missing dir")
	}
}
