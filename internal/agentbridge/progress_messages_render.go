package agentbridge

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func progressArgString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, bool:
		return fmt.Sprint(v)
	default:
		return ""
	}
}

func renderProgressTemplate(template string, args map[string]string) string {
	return progressPlaceholderPattern.ReplaceAllStringFunc(template, func(match string) string {
		parts := progressPlaceholderPattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		value := strings.TrimSpace(args[parts[1]])
		if value == "" {
			value = "not provided"
		}
		return value
	})
}
