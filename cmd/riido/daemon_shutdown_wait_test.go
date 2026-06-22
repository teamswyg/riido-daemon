package main

import (
	"testing"
	"time"
)

func TestWaitDaemonShutdownConditionReturnsWhenDone(t *testing.T) {
	calls := 0
	ok := waitDaemonShutdownCondition(time.Second, func() bool {
		calls++
		return true
	})
	if !ok || calls != 1 {
		t.Fatalf("wait result ok=%v calls=%d", ok, calls)
	}
}

func TestWaitDaemonShutdownConditionHonorsZeroTimeout(t *testing.T) {
	calls := 0
	ok := waitDaemonShutdownCondition(0, func() bool {
		calls++
		return true
	})
	if ok || calls != 0 {
		t.Fatalf("zero-timeout result ok=%v calls=%d", ok, calls)
	}
}
