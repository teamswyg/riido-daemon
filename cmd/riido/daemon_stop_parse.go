package main

import "strconv"

func parseDaemonStopFlags(args []string) (daemonStopFlags, error) {
	flags := daemonStopFlags{timeoutSeconds: 5}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--socket":
			value, next, err := parseDaemonStopValue(args, i, "--socket", "path")
			if err != nil {
				return flags, err
			}
			flags.socket, i = value, next
		case "--pid-file":
			value, next, err := parseDaemonStopValue(args, i, "--pid-file", "path")
			if err != nil {
				return flags, err
			}
			flags.pidFile, i = value, next
		case "--timeout-seconds":
			value, next, err := parseDaemonStopTimeout(args, i)
			if err != nil {
				return flags, err
			}
			flags.timeoutSeconds, i = value, next
		case "--force":
			flags.force = true
		case "--help", "-h":
			printUsage()
			return flags, errCLIHelp
		default:
			return flags, daemonErrorf(ErrDaemonUsage, "stop.parse-flags", "unknown argument: %s", args[i])
		}
	}
	return flags, nil
}

func parseDaemonStopTimeout(args []string, index int) (int, int, error) {
	value, next, err := parseDaemonStopValue(args, index, "--timeout-seconds", "value")
	if err != nil {
		return 0, index, err
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return 0, index, daemonWrapf(ErrDaemonUsage, "stop.parse-flags", err, "--timeout-seconds must be positive int: %v", value)
	}
	return seconds, next, nil
}
