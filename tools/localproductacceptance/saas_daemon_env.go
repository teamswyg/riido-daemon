package main

func saasDaemonEnv(slot saasDaemonSlot, host string) []string {
	return []string{
		"RIIDO_SAAS_URL=" + host,
		"RIIDO_DEVICE_ID=" + slot.Credential.DeviceID,
		"RIIDO_DEVICE_SECRET=" + slot.Credential.DeviceSecret,
		"RIIDO_DEVICE_NAME=" + slot.Credential.DisplayName,
		"RIIDO_DAEMON_PROFILE=staging",
		"RIIDO_DAEMON_PPROF_ADDR=127.0.0.1:0",
		"RIIDO_RUNTIME_OWNER=local-qa",
		"RIIDO_WORKDIR_ROOT=" + slot.Workdir,
		"RIIDO_RUNTIME_MAX_CONCURRENT=2",
		"RIIDO_DAEMON_HEARTBEAT_INTERVAL_SECONDS=3",
		"RIIDO_DAEMON_POLL_INTERVAL_SECONDS=1",
		"RIIDO_DAEMON_IDLE_POLL_INTERVAL_SECONDS=1",
	}
}
