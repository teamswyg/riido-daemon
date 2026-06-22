package netutil

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestApplyContextDeadlineSkipsContextWithoutDeadline(t *testing.T) {
	left, right := net.Pipe()
	defer left.Close()
	defer right.Close()

	if err := ApplyContextDeadline(context.Background(), left, "test pipe"); err != nil {
		t.Fatalf("ApplyContextDeadline() error = %v", err)
	}
}

func TestApplyContextDeadlineAppliesReadTimeout(t *testing.T) {
	left, right := net.Pipe()
	defer left.Close()
	defer right.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	if err := ApplyContextDeadline(ctx, left, "test pipe"); err != nil {
		t.Fatalf("ApplyContextDeadline() error = %v", err)
	}

	var buf [1]byte
	_, err := left.Read(buf[:])
	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Fatalf("Read() error = %v, want timeout", err)
	}
}
