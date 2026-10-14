package lock

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestWithFileSerializesExclusiveLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "task-db.lock")
	first, err := AcquireFile(context.Background(), path)
	if err != nil {
		t.Fatalf("AcquireFile first: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := AcquireFile(ctx, path); err == nil {
		t.Fatal("second acquire should time out while first lock is held")
	}
	if err := first.Release(); err != nil {
		t.Fatalf("Release first: %v", err)
	}
	if err := WithFile(context.Background(), path, func() error { return nil }); err != nil {
		t.Fatalf("WithFile after release: %v", err)
	}
}
