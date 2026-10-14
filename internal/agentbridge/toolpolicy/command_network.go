package toolpolicy

import "strings"

func commandContainsNetworkEgress(command string) bool {
	normalized := strings.ToLower(command)
	return strings.Contains(normalized, "http://") ||
		strings.Contains(normalized, "https://") ||
		strings.Contains(normalized, "curl ") ||
		strings.Contains(normalized, "wget ") ||
		strings.Contains(normalized, "nc ") ||
		strings.Contains(normalized, "netcat ")
}
