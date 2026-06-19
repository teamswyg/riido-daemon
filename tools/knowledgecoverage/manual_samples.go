package main

func manualSamples(docs []docClass, perGroup int) []manualSample {
	seen := map[string]int{}
	samples := []manualSample{}
	for _, doc := range docs {
		if doc.Kind != "manual_registered" || seen[doc.Group] >= perGroup {
			continue
		}
		samples = append(samples, manualSample{
			Group:  doc.Group,
			Path:   doc.Path,
			Reason: doc.Reason,
		})
		seen[doc.Group]++
	}
	return samples
}
