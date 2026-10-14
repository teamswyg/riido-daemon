package codex

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func filterCustomArgs(custom []string) (kept, dropped []string) {
	kept, dropped = agentbridge.FilterBlockedArgs(custom, BlockedArgs())
	for _, rule := range []customArgBlockRule{
		{names: SecurityCriticalArgs(), dropsFollowingValue: true},
		{names: SandboxOverrideArgs(), dropsFollowingValue: true},
		{names: UnsafeBypassArgs()},
	} {
		kept, dropped = filterCustomArgsByRule(kept, dropped, rule)
	}
	return kept, dropped
}

type customArgBlockRule struct {
	names               []string
	dropsFollowingValue bool
}

func filterCustomArgsByRule(custom, dropped []string, rule customArgBlockRule) (kept, allDropped []string) {
	blocked := blockedCustomArgSet(rule.names)
	allDropped = append(allDropped, dropped...)
	for i := 0; i < len(custom); i++ {
		arg := custom[i]
		if _, isBlocked := blocked[arg]; isBlocked {
			allDropped = append(allDropped, arg)
			if rule.dropsFollowingValue && i+1 < len(custom) {
				allDropped = append(allDropped, custom[i+1])
				i++
			}
			continue
		}
		if eq := strings.IndexByte(arg, '='); eq > 0 {
			if _, isBlocked := blocked[arg[:eq]]; isBlocked {
				allDropped = append(allDropped, arg)
				continue
			}
		}
		kept = append(kept, arg)
	}
	return kept, allDropped
}

func blockedCustomArgSet(names []string) map[string]struct{} {
	blocked := make(map[string]struct{}, len(names))
	for _, name := range names {
		blocked[name] = struct{}{}
	}
	return blocked
}
