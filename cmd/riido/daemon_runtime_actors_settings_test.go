package main

func dynamicSaaSRuntimeSettings() daemonSettings {
	return daemonSettings{
		DaemonID:     "daemon-1",
		DeviceName:   "device-1",
		RuntimeOwner: "owner-1",
		SaaSURL:      "https://api.riido.ai",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
		PolicyBundle: "policy-bundle.test.v1",
	}
}
