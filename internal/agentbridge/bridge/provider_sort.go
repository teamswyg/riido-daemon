package bridge

func sortByProvider(caps []RuntimeCapability) {
	// Simple insertion sort keeps Detect stable without an extra dependency.
	for i := 1; i < len(caps); i++ {
		for j := i; j > 0 && caps[j-1].Provider > caps[j].Provider; j-- {
			caps[j-1], caps[j] = caps[j], caps[j-1]
		}
	}
}
