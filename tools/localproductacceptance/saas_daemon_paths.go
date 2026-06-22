package main

import (
	"path/filepath"
	"strconv"
)

func newSaaSDaemonSlot(cfg config, index int, credential saasDeviceCredential) saasDaemonSlot {
	name := "slot-" + strconv.Itoa(index)
	root := filepath.Join(*cfg.daemonRunDir, name)
	return saasDaemonSlot{
		Index:      index,
		Credential: credential,
		Socket:     filepath.Join(root, "riido.sock"),
		PIDFile:    filepath.Join(root, "riido.pid"),
		LogFile:    filepath.Join(root, "riido.log"),
		LockFile:   filepath.Join(root, "riido.lock"),
		Workdir:    filepath.Join(root, "workdir"),
	}
}
