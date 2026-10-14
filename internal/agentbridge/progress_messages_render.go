package agentbridge

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/teamswyg/riido-contracts/progressmessage"
)

func renderProgressMessage(code ProgressCode, args map[string]string) (string, string, bool) {
	args = progressmessage.NormalizeArgsForCode(int(code), args)
	rendered, ok := progressmessage.Render(int(code), args, progressmessage.DefaultLocale)
	if !ok {
		return "", "", false
	}
	return rendered, progressMessageKey(code), true
}

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
