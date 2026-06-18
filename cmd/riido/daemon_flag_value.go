package main

func daemonFlagValue(args []string, index int, flag, valueName string) (string, int, error) {
	next := index + 1
	if next >= len(args) {
		return "", index, daemonErrorf(ErrDaemonUsage, "logs.parse-flags", "%s requires a %s", flag, valueName)
	}
	return args[next], next, nil
}
