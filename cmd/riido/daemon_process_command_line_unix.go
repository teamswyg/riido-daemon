//go:build !windows

package main

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

func daemonProcessCommandLine(pid int) (string, error) {
	var lastErr error
	for _, ps := range []string{"/bin/ps", "/usr/bin/ps", "ps"} {
		out, err := exec.Command(ps, "-p", strconv.Itoa(pid), "-o", "command=").Output()
		if err == nil {
			return strings.TrimSpace(string(bytes.TrimSpace(out))), nil
		}
		lastErr = err
	}
	return "", lastErr
}
