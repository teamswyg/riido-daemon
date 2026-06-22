package main

import "time"

func stopExistingSaaSDaemon(binary string, slot saasDaemonSlot) {
	_, _ = runLocalCommand(5*time.Second, binary,
		"daemon", "stop",
		"--socket", slot.Socket,
		"--pid-file", slot.PIDFile,
		"--timeout-seconds", "2",
		"--force",
	)
}
