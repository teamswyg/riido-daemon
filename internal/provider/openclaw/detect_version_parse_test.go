package openclaw

import "testing"

func TestParseOpenClawVersionRejectsNodeSemver(t *testing.T) {
	cases := []string{
		"requires Node >=22.12.0",
		"node v20.10.0",
		"Node.js v20.10.0",
		"Detected: node 20.10.0 (exec: /usr/bin/node)",
		"package error: dep@22.12.0",
		"    at /path/22.12.0/file.js",
		"22.12.0",
		"20.10.0",
		"v20.10.0",
		"v22.12.0",
		"openclaw requires Node >=22.12.0",
	}
	for _, in := range cases {
		got, ok := parseVersion(in)
		if ok {
			t.Fatalf("parseVersion(%q) MUST reject Node-style semver; got %v", in, got)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	if compareVersions([3]int{2026, 5, 5}, [3]int{2026, 5, 5}) != 0 {
		t.Fatal("equal")
	}
	if compareVersions([3]int{2026, 5, 4}, [3]int{2026, 5, 5}) != -1 {
		t.Fatal("less")
	}
	if compareVersions([3]int{2026, 6, 1}, [3]int{2026, 5, 31}) != 1 {
		t.Fatal("greater")
	}
}
