package main

import (
	"strconv"
	"time"
)

func readSaaSDaemonStatus(binary string, slot saasDaemonSlot) scenario {
	sc := scenario{ID: saasStatusScenarioID(slot.Index), Method: "DAEMON", Endpoint: slot.Socket}
	out, err := runLocalCommand(10*time.Second, binary, "daemon", "status", "--socket", slot.Socket)
	payload := decodeObjectPayload([]byte(out))
	sc.Observed = summarizeDaemonStatus(payload)
	sc.Observed["device_id"] = slot.Credential.DeviceID
	sc.Observed["output_tail"] = outputTail(out)
	if err != nil {
		sc.Status = statusFailed
		sc.FailureSummary = err.Error()
		return sc
	}
	sc.Status = statusPassed
	return sc
}

func summarizeDaemonStatus(payload map[string]any) map[string]any {
	return map[string]any{
		"ready":          payload["ready"],
		"daemon_version": payload["daemon_version"],
		"profile":        payload["profile"],
		"runtimes_count": arrayLen(payload["runtimes"]),
	}
}

func saasStatusScenarioID(slot int) string {
	return "local.saas.daemon_status." + strconv.Itoa(slot)
}
