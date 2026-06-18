package agentbridge

import "strings"

// FilterBlockedArgs removes adapter-blocked args from caller-supplied custom
// args. Both the bare form (--flag value) and equals form (--flag=value) are
// recognized.
func FilterBlockedArgs(custom, blocked []string) (kept, dropped []string) {
	blockedSet := make(map[string]struct{}, len(blocked))
	for _, b := range blocked {
		blockedSet[b] = struct{}{}
	}
	for i := 0; i < len(custom); i++ {
		arg := custom[i]
		if _, isBlocked := blockedSet[arg]; isBlocked {
			dropped = append(dropped, arg)
			if i+1 < len(custom) && !strings.HasPrefix(custom[i+1], "-") {
				dropped = append(dropped, custom[i+1])
				i++
			}
			continue
		}
		if blockedArgWithValue(arg, blockedSet) {
			dropped = append(dropped, arg)
			continue
		}
		kept = append(kept, arg)
	}
	return kept, dropped
}

func blockedArgWithValue(arg string, blockedSet map[string]struct{}) bool {
	eq := strings.IndexByte(arg, '=')
	if eq <= 0 {
		return false
	}
	_, isBlocked := blockedSet[arg[:eq]]
	return isBlocked
}
