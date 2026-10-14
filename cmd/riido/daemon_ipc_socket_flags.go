package main

func requireSocketFlag(args []string) (string, error) {
	if len(args) == 0 {
		return defaultAgentDaemonSocket()
	}
	if isHelpArg(args[0]) {
		printUsage()
		return "", errCLIHelp
	}
	if args[0] != "--socket" {
		return "", daemonErrorf(ErrDaemonUsage, "ipc.parse-socket", "unknown argument: %s", args[0])
	}
	if len(args) < 2 {
		return "", daemonErrorf(ErrDaemonUsage, "ipc.parse-socket", "--socket requires a path")
	}
	if len(args) > 2 {
		return "", daemonErrorf(ErrDaemonUsage, "ipc.parse-socket", "unknown argument: %s", args[2])
	}
	return args[1], nil
}
