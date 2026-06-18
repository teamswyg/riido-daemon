package detectutil

// ProbeResult is the output of VersionProbeStrict. It separates command
// completion from command success.
type ProbeResult struct {
	Output   string
	ExitCode int
	OK       bool
}
