package main

import "sort"

const generatedOriginSampleLimit = 3

func generatedOrigins(docs []docClass) []generatedOrigin {
	byGenerator := map[string]*generatedOrigin{}
	for _, doc := range docs {
		if doc.Kind != "generated" {
			continue
		}
		origin := generatedOriginFor(byGenerator, doc.Generator)
		origin.Count++
		if len(origin.Samples) < generatedOriginSampleLimit {
			origin.Samples = append(origin.Samples, doc.Path)
		}
	}
	return sortedGeneratedOrigins(byGenerator)
}

func generatedOriginFor(byGenerator map[string]*generatedOrigin, generator string) *generatedOrigin {
	if generator == "" {
		generator = "unknown-generated-source"
	}
	if byGenerator[generator] == nil {
		byGenerator[generator] = &generatedOrigin{Generator: generator}
	}
	return byGenerator[generator]
}

func sortedGeneratedOrigins(byGenerator map[string]*generatedOrigin) []generatedOrigin {
	out := make([]generatedOrigin, 0, len(byGenerator))
	for _, origin := range byGenerator {
		out = append(out, *origin)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].Generator < out[j].Generator
		}
		return out[i].Count > out[j].Count
	})
	return out
}
