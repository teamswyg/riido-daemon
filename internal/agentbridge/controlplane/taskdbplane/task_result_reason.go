package taskdbplane

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func resultReason(res agentbridge.Result, fallback string) string {
	return textutil.FirstNonEmptyTrimmed(res.Error, res.Output, fallback)
}
