package main

import (
	"strings"

	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
)

func normalizeCursorRuntimeModelID(modelID string) string {
	if strings.TrimSpace(modelID) == "auto" {
		return providercatalog.DefaultCursorModelID
	}
	return strings.TrimSpace(modelID)
}
