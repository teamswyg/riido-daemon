package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

type levelSource struct {
	Order map[string]int
	Names map[string]string
}

func readSource(repo, rel string) (string, error) {
	path, err := cleanRepoPath(repo, rel)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	return string(data), err
}

func parseLevels(src string) levelSource {
	order := map[string]int{"ShutdownNone": 0, "ShutdownGraceful": 1, "ShutdownForced": 2}
	names := map[string]string{}
	re := regexp.MustCompile(`case (Shutdown\w+):\s*return "([^"]+)"`)
	for _, match := range re.FindAllStringSubmatch(src, -1) {
		names[match[1]] = match[2]
	}
	return levelSource{Order: order, Names: names}
}

func parseTimeouts(src string) map[string]string {
	out := map[string]string{}
	for line := range strings.SplitSeq(src, "\n") {
		parseTimeoutLine(out, strings.TrimSpace(line))
	}
	return out
}

func parseTimeoutLine(out map[string]string, line string) {
	if strings.HasPrefix(line, "DefaultGracefulShutdownTimeout") {
		out["DefaultGracefulShutdownTimeout"] = durationValue(line)
	}
	if strings.HasPrefix(line, "DefaultForcedShutdownTimeout") {
		out["DefaultForcedShutdownTimeout"] = durationValue(line)
	}
}

func durationValue(line string) string {
	if strings.Contains(line, "5 * time.Second") {
		return "5s"
	}
	if strings.Contains(line, "time.Second") {
		return strconv.Itoa(1) + "s"
	}
	return ""
}
