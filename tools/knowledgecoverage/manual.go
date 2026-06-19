package main

func manualPathIndex(m manifest) map[string]manualGroup {
	result := map[string]manualGroup{}
	for _, group := range m.ManualGroups {
		for _, path := range group.Paths {
			result[path] = group
		}
	}
	return result
}

func manualGroupIDs(m manifest) []string {
	ids := make([]string, 0, len(m.ManualGroups))
	for _, group := range m.ManualGroups {
		ids = append(ids, group.ID)
	}
	return ids
}
