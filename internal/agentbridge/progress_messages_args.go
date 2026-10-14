package agentbridge

import "strings"

func cleanProgressArgs(args map[string]any) map[string]string {
	if len(args) == 0 {
		return nil
	}
	out := map[string]string{}
	for key, value := range args {
		key = strings.TrimSpace(key)
		rendered := strings.TrimSpace(progressArgString(value))
		if key != "" && rendered != "" {
			out[key] = rendered
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
