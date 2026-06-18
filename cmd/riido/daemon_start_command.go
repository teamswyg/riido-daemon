package main

import "github.com/teamswyg/riido-daemon/pkg/lifecycle"

func runDaemonStart(ctx lifecycle.Context, args []string) error {
	flags, err := parseStartFlags(args)
	if isCLIHelp(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := applyDefaultStartFlagPaths(&flags); err != nil {
		return err
	}
	if flags.foreground {
		return runDaemonStartForeground(ctx, flags)
	}
	return runDaemonStartBackground(ctx, flags)
}

func applyDefaultStartFlagPaths(flags *startFlags) error {
	if flags.socket == "" {
		def, err := defaultAgentDaemonSocket()
		if err != nil {
			return err
		}
		flags.socket = def
	}
	if flags.lockFile == "" {
		lockPath, err := defaultDaemonLockPath()
		if err != nil {
			return err
		}
		flags.lockFile = lockPath
	}
	return nil
}
