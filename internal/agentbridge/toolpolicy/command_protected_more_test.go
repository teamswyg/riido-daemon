package toolpolicy

import "testing"

func TestCommandTouchesProtectedPathCanonicalizesCommandPaths(t *testing.T) {
	for _, command := range []string{
		`printf token > .docker\config.json`,
		`cp token repo\\.config\\gh\\hosts.yml`,
		`sed -i s/a/b/ repo//.ssh//config`,
	} {
		if !commandTouchesProtectedPath(command) {
			t.Fatalf("command should touch protected path: %q", command)
		}
	}
}

func TestCommandTouchesProtectedPathRequiresWriteMarker(t *testing.T) {
	for _, command := range []string{
		`echo .docker\config.json`,
		`printf .config/gh/hosts.yml`,
		`ls repo//.ssh//config`,
	} {
		if commandTouchesProtectedPath(command) {
			t.Fatalf("read-only mention should not be protected write: %q", command)
		}
	}
}
