package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "riido:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 1 {
		printUsage()
		return fmt.Errorf("missing command")
	}
	if args[0] == "--help" || args[0] == "-h" {
		printUsage()
		return nil
	}
	switch args[0] {
	case "mwsd":
		return runMwsd(args[1:])
	case "task":
		return runTask(args[1:])
	case "serve":
		return runServe(args[1:])
	case "api":
		return runAPI(args[1:])
	case "bridge":
		return runBridge(args[1:])
	case "daemon":
		return runDaemon(args[1:])
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", args[0])
	}
}
