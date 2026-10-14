package main

type daemonStopFlags struct {
	socket         string
	pidFile        string
	timeoutSeconds int
	force          bool
}

func parseDaemonStopValue(args []string, index int, flag, valueName string) (string, int, error) {
	next := index + 1
	if next >= len(args) {
		return "", index, daemonErrorf(ErrDaemonUsage, "stop.parse-flags", "%s requires a %s", flag, valueName)
	}
	return args[next], next, nil
}
