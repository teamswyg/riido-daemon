package main

import "testing"

func TestGoStringSliceLiteral(t *testing.T) {
	got := goStringSliceLiteral([]string{"alpha", "beta"})
	if got != `[]string{"alpha", "beta"}` {
		t.Fatalf("unexpected literal: %s", got)
	}
}

func TestGoStringSliceLiteralForEmptySlice(t *testing.T) {
	if got := goStringSliceLiteral(nil); got != "nil" {
		t.Fatalf("unexpected empty literal: %s", got)
	}
}
