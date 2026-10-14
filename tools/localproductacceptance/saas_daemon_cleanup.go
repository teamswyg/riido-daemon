package main

import "time"

func cleanupSaaSDaemons(binary string, prep saasPrepareResult) []scenario {
	out := make([]scenario, 0, len(prep.Slots))
	for _, slot := range prep.Slots {
		out = append(out, stopSaaSDaemonScenario(binary, slot))
	}
	return out
}

func stopSaaSDaemonScenario(binary string, slot saasDaemonSlot) scenario {
	sc := scenario{ID: "local.saas.daemon_cleanup." + intString(slot.Index), Method: "DAEMON", Endpoint: slot.Socket}
	out, err := runLocalCommand(10*time.Second, binary,
		"daemon", "stop",
		"--socket", slot.Socket,
		"--pid-file", slot.PIDFile,
		"--timeout-seconds", "3",
		"--force",
	)
	sc.Observed = map[string]any{"device_id": slot.Credential.DeviceID, "output_tail": outputTail(out)}
	if err != nil {
		sc.Status = statusFailed
		sc.FailureSummary = err.Error()
		return sc
	}
	sc.Status = statusPassed
	return sc
}
