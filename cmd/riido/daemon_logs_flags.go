package main

import "strconv"

func parseDaemonLogsFlags(args []string) (daemonLogsFlags, bool, error) {
	flags := daemonLogsFlags{lines: 50}
	for i := 0; i < len(args); i++ {
		next, ok, err := parseDaemonLogsFlag(args, i, &flags)
		if err != nil || !ok {
			return flags, false, err
		}
		i = next
	}
	if flags.logFile == "" {
		return flags, false, daemonErrorf(ErrDaemonUsage, "logs.parse-flags", "--log-file is required")
	}
	return flags, true, nil
}

func parseDaemonLogsFlag(args []string, index int, flags *daemonLogsFlags) (int, bool, error) {
	switch args[index] {
	case "--log-file":
		value, next, err := daemonFlagValue(args, index, "--log-file", "path")
		flags.logFile = value
		return next, true, err
	case "--lines":
		return parseDaemonLogsLinesFlag(args, index, flags)
	case "--help", "-h":
		printUsage()
		return index, false, nil
	default:
		return index, false, daemonErrorf(ErrDaemonUsage, "logs.parse-flags", "unknown argument: %s", args[index])
	}
}

func parseDaemonLogsLinesFlag(args []string, index int, flags *daemonLogsFlags) (int, bool, error) {
	value, next, err := daemonFlagValue(args, index, "--lines", "value")
	if err != nil {
		return next, false, err
	}
	lines, err := strconv.Atoi(value)
	if err != nil || lines <= 0 {
		return next, false, daemonWrapf(ErrDaemonUsage, "logs.parse-flags", err, "--lines must be positive int")
	}
	flags.lines = lines
	return next, true, nil
}
