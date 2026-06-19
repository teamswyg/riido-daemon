package main

func manifestBehaviors(m Manifest) []string {
	seen := map[string]bool{}
	var out []string
	add := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	for _, row := range m.SearchOrder {
		add(row.Behavior)
	}
	for _, row := range m.Rules {
		add(row.Behavior)
	}
	return out
}
