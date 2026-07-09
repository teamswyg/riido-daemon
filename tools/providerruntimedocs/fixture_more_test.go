package main

func testManifest() manifest {
	return manifest{
		SchemaVersion:    manifestSchema,
		ID:               "provider-runtime-docs-test",
		Title:            "Provider Runtime Test",
		GeneratedDoc:     "docs/provider-runtime.md",
		Workflow:         ".github/workflows/provider-runtime-docs.yml",
		EvidenceArtifact: "out/provider-runtime-docs.json",
		Summary:          []string{"Provider runtime docs are generated."},
		CompatibilityMarkers: []string{
			"runtime snapshot",
			"provider adapter",
		},
		Parts: []link{
			{Title: "one", Path: "one.md"},
			{Title: "two", Path: "two.md"},
			{Title: "three", Path: "three.md"},
			{Title: "four", Path: "four.md"},
			{Title: "five", Path: "five.md"},
			{Title: "six", Path: "six.md"},
			{Title: "seven", Path: "seven.md"},
			{Title: "eight", Path: "eight.md"},
		},
		RelatedPages: []link{{Title: "evidence", Path: "evidence.md"}},
		Assertions:   []string{"Runtime docs are executable knowledge."},
		Pages: []page{
			testPage("claude"),
			testPage("codex"),
			testPage("cursor"),
		},
	}
}

func testPage(id string) page {
	return page{
		SchemaVersion: pageSchema,
		ID:            id,
		Title:         "Runtime " + id,
		GeneratedDoc:  "docs/provider-runtime/" + id + ".md",
		BackTitle:     "root",
		BackPath:      "../provider-runtime.md",
		Blocks: []block{
			{Kind: "heading", Text: "Boundary"},
			{Kind: "paragraph", Text: "Provider adapter boundary."},
			{Kind: "bullets", Items: []string{"detect", "execute"}},
			{Kind: "links", Links: []link{{Title: "source", Path: "source.go"}}},
		},
	}
}
