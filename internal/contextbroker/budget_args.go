package contextbroker

import (
	"strconv"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// budgetChars reads `--budget CHARS` (sent as budget_chars); 0 means none.
func budgetChars(args map[string]string) int {
	n, err := strconv.Atoi(args["budget_chars"])
	if err != nil || n < 1 {
		return 0
	}
	return n
}

func withBudgetWarnings(env contextcli.Envelope, cut, fits bool) contextcli.Envelope {
	if cut {
		env.Warnings = append(env.Warnings, contextcli.Warning{Code: "BUDGET_TRUNCATED", Message: "--budget에 맞추려고 긴 본문을 줄였습니다(…로 표시)"})
	}
	if !fits {
		env.Warnings = append(env.Warnings, contextcli.Warning{Code: "BUDGET_EXCEEDED", Message: "본문을 줄여도 --budget보다 깁니다. 더 좁은 명령(riido work get)을 쓰세요"})
	}
	return env
}
