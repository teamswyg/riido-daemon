package runtimeactor

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
)

func detectedFingerprintForExecutable(executable string) providercap.DetectedFingerprint {
	executable = strings.TrimSpace(executable)
	if executable == "" || !filepath.IsAbs(executable) {
		return ""
	}
	info, err := os.Stat(executable)
	if err != nil || !info.Mode().IsRegular() {
		return ""
	}
	return fingerprintRegularFile(executable)
}

func fingerprintRegularFile(executable string) providercap.DetectedFingerprint {
	file, err := os.Open(executable)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}
	return providercap.DetectedFingerprint(hex.EncodeToString(hash.Sum(nil)))
}
