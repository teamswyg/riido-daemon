package main

import (
	"os"
	"path/filepath"
	"strings"
)

const manifestLoopSampleLimit = 3

func scanManifestLoops(root string) (manifestLoopReport, error) {
	report := manifestLoopReport{}
	counts := map[string]int{}
	samples := map[string][]string{}
	delegatedTargets, err := scanManifestLoopDelegatedTargets(root)
	if err != nil {
		return report, err
	}
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && filepath.Base(path) == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".riido.json") {
			return nil
		}
		scanManifestLoopPath(root, path, delegatedTargets, &report, counts, samples)
		return nil
	})
	report.MissingGroups = manifestGroups(counts)
	report.MissingSamples = manifestSamples(report.MissingGroups, samples)
	return report, err
}

func scanManifestLoopPath(
	root string,
	path string,
	delegatedTargets map[string]bool,
	report *manifestLoopReport,
	counts map[string]int,
	samples map[string][]string,
) {
	group := manifestGroup(root, path)
	status := manifestLoopStatus(root, path)
	if status == "missing" && delegatedTargets[path] {
		status = "delegated"
	}
	switch status {
	case "direct":
		report.Complete++
		report.Direct++
	case "delegated":
		report.Complete++
		report.Delegated++
	default:
		report.Missing++
		counts[group]++
		if len(samples[group]) < manifestLoopSampleLimit {
			samples[group] = append(samples[group], slashPath(root, path))
		}
	}
}
