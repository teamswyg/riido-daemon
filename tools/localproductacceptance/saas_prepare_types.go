package main

type saasPrepareResult struct {
	Scenarios []scenario
	DeviceIDs []string
	Runtimes  []preparedRuntime
	Slots     []saasDaemonSlot
}

type saasDeviceCredential struct {
	DeviceID     string
	DeviceSecret string
	DisplayName  string
}

type saasDaemonSlot struct {
	Index      int
	Credential saasDeviceCredential
	Socket     string
	PIDFile    string
	LogFile    string
	LockFile   string
	Workdir    string
}

type preparedRuntime struct {
	DeviceID        string
	RuntimeID       string
	Kind            string
	ProviderVersion string
}
