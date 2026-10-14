package main

func prepareSaaSDaemonSlot(cfg config, client apiClient, index int) (saasDaemonSlot, []scenario) {
	credential, enroll := enrollSaaSDevice(cfg, client, index)
	slot := newSaaSDaemonSlot(cfg, index, credential)
	scenarios := []scenario{enroll}
	if enroll.Status != statusPassed {
		return slot, scenarios
	}
	stopExistingSaaSDaemon(*cfg.daemonBinary, slot)
	start := startSaaSDaemon(*cfg.daemonBinary, slot, *cfg.agentHost)
	scenarios = append(scenarios, start)
	if start.Status != statusPassed {
		return slot, scenarios
	}
	status := readSaaSDaemonStatus(*cfg.daemonBinary, slot)
	scenarios = append(scenarios, status)
	return slot, scenarios
}
