package main

import "os"

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func remove(values []string, target string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value != target {
			out = append(out, value)
		}
	}
	return out
}
