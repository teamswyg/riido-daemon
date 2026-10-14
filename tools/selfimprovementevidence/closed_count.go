package main

func countVerifiedClosed(loops []closedSummary) int {
	count := 0
	for _, item := range loops {
		if item.Status == statusVerified {
			count++
		}
	}
	return count
}
