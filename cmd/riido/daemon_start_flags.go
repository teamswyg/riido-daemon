package main

type startFlags struct {
	foreground bool
	socket     string
	pidFile    string
	logFile    string
	lockFile   string
}

func parseStartFlags(args []string) (startFlags, error) {
	out := startFlags{}
	for i := 0; i < len(args); i++ {
		var err error
		i, err = parseStartFlagAt(args, i, &out)
		if err != nil {
			return out, err
		}
	}
	return out, nil
}

func parseStartFlagAt(args []string, i int, out *startFlags) (int, error) {
	switch args[i] {
	case "--foreground":
		out.foreground = true
	case "--socket":
		return parseStartPathFlag(args, i, "--socket", &out.socket)
	case "--pid-file":
		return parseStartPathFlag(args, i, "--pid-file", &out.pidFile)
	case "--log-file":
		return parseStartPathFlag(args, i, "--log-file", &out.logFile)
	case "--lock-file":
		return parseStartPathFlag(args, i, "--lock-file", &out.lockFile)
	case "--help", "-h":
		printUsage()
		return i, errCLIHelp
	default:
		return i, daemonErrorf(ErrDaemonUsage, "start.parse-flags", "unknown argument: %s", args[i])
	}
	return i, nil
}

func parseStartPathFlag(args []string, i int, flag string, target *string) (int, error) {
	i++
	if i >= len(args) {
		return i, daemonErrorf(ErrDaemonUsage, "start.parse-flags", "%s requires a path", flag)
	}
	*target = args[i]
	return i, nil
}
