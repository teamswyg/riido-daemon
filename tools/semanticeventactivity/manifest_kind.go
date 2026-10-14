package main

func manifestKindMap(manifest Manifest) (map[string]bool, []problem) {
	kinds := map[string]bool{}
	var problems []problem
	add := func(kind string, semantic bool) {
		if kind == "" {
			problems = append(problems, problem{"empty event kind"})
			return
		}
		if previous, ok := kinds[kind]; ok {
			problems = append(problems, problem{"duplicate event kind: " + kind})
			if previous != semantic {
				problems = append(problems, problem{"overlapping semantic categories: " + kind})
			}
			return
		}
		kinds[kind] = semantic
	}
	for _, kind := range manifest.SemanticActivity {
		add(kind, true)
	}
	for _, kind := range manifest.NonSemanticActivity {
		add(kind, false)
	}
	return kinds, problems
}
