package detectutil

// augmentedSearchDirs is the seam tests override; production resolution uses
// productionSearchDirs.
var augmentedSearchDirs = productionSearchDirs

// productionSearchDirs returns the ordered, de-duplicated directories to scan
// for an executable.
func productionSearchDirs() []string {
	collector := newSearchDirCollector()
	collector.addSplitPath(processPATH())
	collector.addDirs(loginShellPATHDirs())
	collector.addDirs(wellKnownInstallDirs())
	return collector.values
}
