package detectutil

import (
	"os/exec"
	"path/filepath"
	"strings"
)

func pathExecutableCandidates(name string) []string {
	collector := newCandidateCollector()
	if path, err := exec.LookPath(name); err == nil {
		collector.add(path)
	}
	if strings.ContainsAny(name, `/\`) {
		return collector.values
	}
	for _, dir := range augmentedSearchDirs() {
		if dir == "" {
			dir = "."
		}
		addExecutableCandidatesFromDir(&collector, dir, name)
	}
	return collector.values
}

func addExecutableCandidatesFromDir(collector *candidateCollector, dir, name string) {
	for _, candidateName := range executableNames(name) {
		path := filepath.Join(dir, candidateName)
		if isExecutableFile(path) {
			collector.add(path)
		}
	}
}

type candidateCollector struct {
	values []string
	seen   map[string]struct{}
}

func newCandidateCollector() candidateCollector {
	return candidateCollector{seen: map[string]struct{}{}}
}

func (c *candidateCollector) add(path string) {
	if path == "" {
		return
	}
	key := filepath.Clean(path)
	if _, ok := c.seen[key]; ok {
		return
	}
	c.seen[key] = struct{}{}
	c.values = append(c.values, path)
}
