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
	if args[0] == "--version" || args[0] == "-v" {
		return runVersion(nil)
	}
	switch mainCommand(args[0]) {
	case mainCommandMwsd:
		return runMwsd(args[1:])
	case mainCommandTask:
		return runTask(args[1:])
	case mainCommandServe:
		return runServe(args[1:])
	case mainCommandAPI:
		return runAPI(args[1:])
	case mainCommandBridge:
		return runBridge(args[1:])
	case mainCommandClaudePermissionMCP:
		return runClaudePermissionMCP(args[1:])
	case mainCommandDaemon:
		return runDaemon(args[1:])
	case mainCommandVersion:
		return runVersion(args[1:])
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", args[0])
	}
}
