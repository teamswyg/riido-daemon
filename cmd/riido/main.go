package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		var reported reportedExit
		if errors.As(err, &reported) {
			os.Exit(reported.ExitCode())
		}
		fmt.Fprintln(os.Stderr, "riido:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 1 {
		printRootUsage()
		return fmt.Errorf("missing command")
	}
	if args[0] == "--help" || args[0] == "-h" {
		printRootUsage()
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
	case mainCommandWork:
		return runWork(defaultContextIO(), args[1:])
	case mainCommandThread:
		return runThread(defaultContextIO(), args[1:])
	case mainCommandRepo:
		return runRepo(defaultContextIO(), args[1:])
	case mainCommandResource:
		return runResource(defaultContextIO(), args[1:])
	case mainCommandContext:
		return runContext(defaultContextIO(), args[1:])
	case mainCommandAgents:
		return runAgents(defaultContextIO(), args[1:])
	case mainCommandMeeting:
		return runMeeting(defaultContextIO(), args[1:])
	case mainCommandAuth:
		return runAuth(defaultContextIO(), args[1:])
	case mainCommandWhoami:
		return runWhoami(defaultContextIO(), args[1:])
	case mainCommandDoctor:
		return runDoctor(defaultContextIO(), args[1:])
	case mainCommandCapabilities:
		return runCapabilities(defaultContextIO(), args[1:])
	default:
		printRootUsage()
		return fmt.Errorf("unknown command: %s", args[0])
	}
}
