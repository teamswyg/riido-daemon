package openclaw

import (
	"context"
	"strconv"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

// Detect resolves the openclaw executable and inspects `openclaw --version`.
func Detect(ctx context.Context, env agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	candidates := detectutil.ResolveExecutableCandidates(DefaultExecutable, envValue(env, EnvOverride))
	if len(candidates) == 0 {
		return agentbridge.DetectResult{
			Available: false,
			Reason:    "openclaw executable not found on PATH and " + EnvOverride + " is not set",
		}, nil
	}

	var first agentbridge.DetectResult
	for i, exe := range candidates {
		res := detectExecutable(ctx, exe)
		if len(candidates) > 1 {
			res.Metadata["path_candidate_count"] = strconv.Itoa(len(candidates))
			res.Metadata["path_candidate_index"] = strconv.Itoa(i + 1)
		}
		if i == 0 {
			first = res
		}
		if res.Available {
			return res, nil
		}
	}

	return first, nil
}
