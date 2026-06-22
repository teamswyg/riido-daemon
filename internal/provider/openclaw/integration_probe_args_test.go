package openclaw

func configProbeArgs() []string {
	args := []string{
		"agent",
		"--local",
		"--json",
		"--session-id",
		"riido-config-probe",
		"--message",
		"Say OK only.",
		"--timeout",
		"30",
	}
	return args
}
