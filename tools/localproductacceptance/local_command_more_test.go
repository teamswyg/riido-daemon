package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRunLocalCommandReportsOutputAndErrors(t *testing.T) {
	out, err := runLocalCommand(time.Second, "sh", "-c", "printf ok")
	if err != nil || out != "ok" {
		t.Fatalf("success out=%q err=%v", out, err)
	}
	out, err = runLocalCommand(time.Second, "sh", "-c", "printf nope; exit 7")
	if err == nil || out != "nope" {
		t.Fatalf("failure out=%q err=%v", out, err)
	}
	out, err = runLocalCommand(time.Nanosecond, "sh", "-c", "sleep 1")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout out=%q err=%v", out, err)
	}
}

func TestOutputTailKeepsShortOutputAndTrimsLongOutput(t *testing.T) {
	if got := outputTail("  short\n"); got != "short" {
		t.Fatalf("short tail=%q", got)
	}
	long := strings.Repeat("a", 700)
	got := outputTail(long)
	if len(got) != 600 || got != strings.Repeat("a", 600) {
		t.Fatalf("long tail len=%d", len(got))
	}
}
