package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/user"
	"strconv"
)

func localQAMachineID(slot int) string {
	host, _ := os.Hostname()
	current, _ := user.Current()
	uid := ""
	if current != nil {
		uid = current.Uid
	}
	sum := sha256.Sum256([]byte(host + "|" + uid + "|" + strconv.Itoa(slot)))
	return "riido-local-qa-" + hex.EncodeToString(sum[:12]) + "-" + strconv.Itoa(slot)
}
