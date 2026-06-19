package main

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

func verifySpawnLaunchPath() error {
	env := detectutil.EnvListWithLaunchPATHFromMap(
		[]string{"RIIDO_TEST=1"},
		map[string]string{"PATH": "/frozen/bin"},
	)
	if pathValue(env) != "/frozen/bin" {
		return behaviorError("spawn env did not receive frozen launch PATH")
	}
	return nil
}

func verifySpawnExplicitPath() error {
	env := detectutil.EnvListWithLaunchPATHFromMap(
		[]string{"PATH=/spawn/bin"},
		map[string]string{"PATH": "/frozen/bin"},
	)
	if pathValue(env) != "/spawn/bin" {
		return behaviorError("explicit spawn PATH was overwritten")
	}
	return nil
}

func pathValue(env []string) string {
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if ok && strings.EqualFold(key, "PATH") {
			return value
		}
	}
	return ""
}
