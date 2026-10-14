package main

import (
	"context"
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
)

func runMwsd(args []string) error {
	if len(args) < 1 {
		printUsage()
		return fmt.Errorf("missing mwsd command")
	}
	if mwsdHelpArg(args[0]) {
		printUsage()
		return nil
	}
	options, err := parseMwsdOptions(args)
	if err != nil {
		return err
	}
	if options.showUsage {
		printUsage()
		return nil
	}
	if err := options.applyDefaults(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client := mwsdbridge.NewClient(options.socketPath)
	return runMwsdCommand(ctx, client, options)
}

func runMwsdCommand(ctx context.Context, client mwsdbridge.Client, options mwsdOptions) error {
	switch options.command {
	case mwsdCommandSnapshot:
		return printMwsdSnapshot(ctx, client)
	case mwsdCommandProjection:
		return printMwsdProjection(ctx, client)
	case mwsdCommandSync:
		return runMwsdSync(ctx, client, options)
	case mwsdCommandOrchestration:
		return printMwsdMethod[mwsdbridge.OrchestrationSnapshot](ctx, client, mwsdbridge.MethodOrchestration)
	case mwsdCommandProjects:
		return printMwsdMethod[mwsdbridge.ProjectRegistry](ctx, client, mwsdbridge.MethodProjects)
	case mwsdCommandStatus:
		return printMwsdMethod[mwsdbridge.Status](ctx, client, mwsdbridge.MethodStatus)
	default:
		printUsage()
		return fmt.Errorf("unknown mwsd command: %s", options.command)
	}
}
