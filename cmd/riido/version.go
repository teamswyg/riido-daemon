package main

import (
	"fmt"
	"strings"
)

var binaryVersion = "dev"

func runVersion(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("version takes no arguments")
	}
	fmt.Println(versionLabel())
	return nil
}

func versionLabel() string {
	version := strings.TrimSpace(binaryVersion)
	if version == "" {
		version = "dev"
	}
	return "riido-daemon " + version
}

func agentDaemonVersionLabel() string {
	version := strings.TrimSpace(binaryVersion)
	if version == "" {
		version = "dev"
	}
	return "riido-agentd " + version
}
