package detectutil

import (
	"path/filepath"
	"strings"
)

type searchDirCollector struct {
	values []string
	seen   map[string]struct{}
}

func newSearchDirCollector() searchDirCollector {
	return searchDirCollector{seen: map[string]struct{}{}}
}

func (c *searchDirCollector) addSplitPath(path string) {
	c.addDirs(filepath.SplitList(path))
}

func (c *searchDirCollector) addDirs(dirs []string) {
	for _, dir := range dirs {
		c.add(dir)
	}
}

func (c *searchDirCollector) add(dir string) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return
	}
	key := filepath.Clean(dir)
	if _, ok := c.seen[key]; ok {
		return
	}
	c.seen[key] = struct{}{}
	c.values = append(c.values, dir)
}
