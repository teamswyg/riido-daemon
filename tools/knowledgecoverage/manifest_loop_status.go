package main

func manifestLoopStatus(root, path string) string {
	return manifestLoopStatusSeen(root, path, map[string]bool{})
}

func manifestLoopStatusSeen(root, path string, seen map[string]bool) string {
	if seen[path] {
		return "missing"
	}
	seen[path] = true
	if manifestDocHasLoop(path) {
		return "direct"
	}
	if source, ok := manifestLoopSourcePath(root, path); ok {
		if manifestLoopStatusSeen(root, source, seen) != "missing" {
			return "delegated"
		}
	}
	return "missing"
}
