package main

import (
	"fmt"
	"strings"
)

func parseLaunchdPrint(out string) launchdEvidence {
	return launchdEvidence{
		State:           parseLaunchdString(out, "state = "),
		Runs:            parseLaunchdInt(out, "runs = "),
		LastExitCode:    parseLaunchdString(out, "last exit code = "),
		CalendarTrigger: strings.Contains(out, "com.apple.launchd.calendarinterval"),
	}
}

func parseLaunchdString(out, prefix string) string {
	for line := range strings.SplitSeq(out, "\n") {
		text := strings.TrimSpace(line)
		if value, ok := strings.CutPrefix(text, prefix); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func parseLaunchdInt(out, prefix string) int {
	value := parseLaunchdString(out, prefix)
	var n int
	_, err := fmt.Sscanf(value, "%d", &n)
	if err != nil {
		return 0
	}
	return n
}
