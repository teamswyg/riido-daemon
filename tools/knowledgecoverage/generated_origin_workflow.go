package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generatedOriginWorkflowProblems(root string, origins []generatedOrigin) []string {
	coverage := scanGeneratedOriginWorkflowCoverage(root, origins)
	var problems []string
	for _, missing := range coverage.Missing {
		problems = append(problems, fmt.Sprintf(
			"generated origin %q has no workflow reference for %s",
			missing.Generator,
			missing.Tool,
		))
	}
	return problems
}

func scanGeneratedOriginWorkflowCoverage(
	root string,
	origins []generatedOrigin,
) generatedOriginWorkflowCoverage {
	workflowText := readWorkflowText(filepath.Join(root, ".github", "workflows"))
	coverage := generatedOriginWorkflowCoverage{Missing: []generatedOriginWorkflowMiss{}}
	for _, origin := range origins {
		tool, ok := generatedToolPath(origin.Generator)
		if !ok {
			continue
		}
		if strings.Contains(workflowText, tool) {
			coverage.CoveredCount++
			continue
		}
		coverage.MissingCount++
		coverage.Missing = append(coverage.Missing, generatedOriginWorkflowMiss{
			Generator: origin.Generator,
			Tool:      tool,
			Count:     origin.Count,
		})
	}
	return coverage
}

func readWorkflowText(dir string) string {
	var b strings.Builder
	_ = filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !isWorkflowFile(path) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err == nil {
			b.Write(data)
			b.WriteByte('\n')
		}
		return nil
	})
	return b.String()
}
