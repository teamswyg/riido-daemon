package codex

import "testing"

func assertArgPair(t *testing.T, args []string, key, value string) {
	t.Helper()
	for i := 0; i+1 < len(args); i++ {
		if args[i] == key && args[i+1] == value {
			return
		}
	}
	t.Fatalf("missing arg pair %s %s in %v", key, value, args)
}

func assertArgBefore(t *testing.T, args []string, before, after string) {
	t.Helper()
	beforeIndex := -1
	afterIndex := -1
	for i, arg := range args {
		if arg == before && beforeIndex == -1 {
			beforeIndex = i
		}
		if arg == after && afterIndex == -1 {
			afterIndex = i
		}
	}
	if beforeIndex == -1 || afterIndex == -1 || beforeIndex >= afterIndex {
		t.Fatalf("expected %q before %q in %v", before, after, args)
	}
}

func assertArgCount(t *testing.T, args []string, key string, want int) {
	t.Helper()
	got := 0
	for _, arg := range args {
		if arg == key {
			got++
		}
	}
	if got != want {
		t.Fatalf("arg %q count = %d, want %d in %v", key, got, want, args)
	}
}
