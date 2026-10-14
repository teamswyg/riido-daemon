package main

func generatedDocPaths(m manifest) []string {
	return []string{
		m.GeneratedDoc,
		m.InvariantsIndex.GeneratedDoc,
		m.Core.GeneratedDoc,
	}
}
