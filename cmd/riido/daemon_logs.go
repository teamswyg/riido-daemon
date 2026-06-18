package main

type daemonLogsFlags struct {
	logFile string
	lines   int
}

func runDaemonLogs(args []string) error {
	flags, ok, err := parseDaemonLogsFlags(args)
	if err != nil || !ok {
		return err
	}
	return printDaemonLogTail(flags)
}
