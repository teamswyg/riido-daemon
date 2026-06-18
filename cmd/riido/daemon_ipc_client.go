package main

func runDaemonStatus(args []string) error {
	return runDaemonSocketCommand(args, daemonMethodStatus)
}

func runDaemonHealth(args []string) error {
	return runDaemonSocketCommand(args, daemonMethodHealth)
}

func runDaemonReady(args []string) error {
	return runDaemonSocketCommand(args, daemonMethodReady)
}

func runDaemonMetrics(args []string) error {
	return runDaemonSocketCommand(args, daemonMethodMetrics)
}

func runDaemonSocketCommand(args []string, method daemonMethod) error {
	sock, err := requireSocketFlag(args)
	if isCLIHelp(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return daemonCall(sock, method)
}
