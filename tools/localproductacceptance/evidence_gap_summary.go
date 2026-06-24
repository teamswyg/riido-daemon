package main

import "strings"

func summarizeEvidenceGaps(items []scenario) evidenceGapSummary {
	out := evidenceGapSummary{}
	for _, item := range items {
		if item.Status == statusSkipped {
			out.Skipped = append(out.Skipped, item.ID)
		}
		if strings.HasPrefix(item.ID, "figma.") && item.Screenshot == "" {
			out.FigmaWithoutScreenshot = append(out.FigmaWithoutScreenshot, item.ID)
		}
	}
	return out
}
