package agentbridge

import (
	"strconv"
	"strings"
)

func ProgressEventMetadata(ev Event) map[string]string {
	if ev.ProgressCode <= 0 {
		return nil
	}
	metadata := map[string]string{
		ProgressMessageMetadataCode: strconv.Itoa(int(ev.ProgressCode)),
	}
	if strings.TrimSpace(ev.ProgressKey) != "" {
		metadata[ProgressMessageMetadataKey] = strings.TrimSpace(ev.ProgressKey)
	}
	for key, value := range ev.ProgressArgs {
		addProgressArgMetadata(metadata, key, value)
	}
	return metadata
}

func addProgressArgMetadata(metadata map[string]string, key, value string) {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" || value == "" {
		return
	}
	metadata[ProgressMessageMetadataArgPrefix+key] = value
}
