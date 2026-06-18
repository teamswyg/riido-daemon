package openclaw

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

func sessionIDFromTaskID(taskID string) string {
	taskID = strings.TrimSpace(taskID)
	if isOpenClawSessionID(taskID) {
		return taskID
	}

	var b strings.Builder
	lastDash := false
	for _, r := range taskID {
		if isOpenClawSessionRune(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	slug := strings.Trim(b.String(), "-_")
	if slug == "" {
		slug = "task"
	}
	return openClawSessionIDWithHash(taskID, slug)
}

func openClawSessionIDWithHash(taskID, slug string) string {
	sum := sha256.Sum256([]byte(taskID))
	hash := fmt.Sprintf("%x", sum[:6])
	const maxSessionIDLen = 80
	maxSlugLen := maxSessionIDLen - len("riido--") - len(hash)
	if len(slug) > maxSlugLen {
		slug = strings.Trim(slug[:maxSlugLen], "-_")
		if slug == "" {
			slug = "task"
		}
	}
	return fmt.Sprintf("riido-%s-%s", slug, hash)
}
