package lock

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestAcquireFileRejectsEmptyPath(t *testing.T) {
	_, err := AcquireFile(context.Background(), " ")
	if err == nil || !strings.Contains(err.Error(), "empty path") {
		t.Fatalf("expected empty path error, got %v", err)
	}
}

func TestReleaseNilAndDoubleReleaseAreNoops(t *testing.T) {
	var nilLock *FileLock
	if err := nilLock.Release(); err != nil {
		t.Fatalf("nil release: %v", err)
	}
	path := filepath.Join(t.TempDir(), "riido.lock")
	lock, err := AcquireFile(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Release(); err != nil {
		t.Fatal(err)
	}
	if err := lock.Release(); err != nil {
		t.Fatalf("double release: %v", err)
	}
	next, err := AcquireFile(context.Background(), path)
	if err != nil {
		t.Fatalf("released lock should be acquirable: %v", err)
	}
	if err := next.Release(); err != nil {
		t.Fatal(err)
	}
}

func TestWithFileReturnsCallbackError(t *testing.T) {
	want := errors.New("callback failed")
	path := filepath.Join(t.TempDir(), "riido.lock")
	err := WithFile(context.Background(), path, func() error { return want })
	if !errors.Is(err, want) {
		t.Fatalf("WithFile error = %v, want %v", err, want)
	}
}
