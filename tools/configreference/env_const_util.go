package main

func manifestEnvNames(manifest Manifest) []string {
	out := make([]string, 0, len(manifest.DaemonEnvVars))
	for _, envVar := range manifest.DaemonEnvVars {
		out = append(out, envVar.Name)
	}
	return out
}

func mapValues(values map[string]string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}
