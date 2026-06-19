package main

import (
	"fmt"
)

func runBridge(args []string) error {
	if len(args) < 1 {
		printUsage()
		return fmt.Errorf("missing bridge subcommand")
	}
	switch bridgeCommand(args[0]) {
	case bridgeCommandProviders:
		return runBridgeProviders(args[1:])
	case bridgeCommandDetect:
		return runBridgeDetect(args[1:])
	default:
		printUsage()
		return fmt.Errorf("unknown bridge subcommand: %s", args[0])
	}
}
