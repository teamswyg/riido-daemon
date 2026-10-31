package main

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func runThread(c contextIO, raw []string) error {
	requestID := contextcli.NewRequestID()
	if len(raw) == 0 || isHelpArg(raw[0]) {
		printThreadUsage(c.stderr)
		if len(raw) == 0 {
			return c.fail(false, requestID, contextcli.Fail(contextcli.CodeUsage, "thread 하위 명령이 필요합니다", "riido thread --help"))
		}
		return nil
	}
	if raw[0] != "list" && raw[0] != "show" {
		return c.fail(wantsJSON(raw), requestID, contextcli.Fail(contextcli.CodeUsage, "지원하지 않는 thread 하위 명령입니다", "riido thread --help"))
	}
	maxPositionals := 0
	if raw[0] == "show" {
		maxPositionals = 1
	}
	args, cliErr := contextcli.ParseArgs(raw[1:], contextcli.ArgSpec{MaxPositionals: maxPositionals, Values: []string{"--work", "--page", "--budget"}})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printThreadUsage(c.stderr)
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "thread 조회는 사용자 모드에서만 지원합니다", "사용자 터미널에서 워크스페이스·작업 ID를 지정하세요"))
	}
	if slices.ContainsFunc([]string{"--cursor", "--page", "--budget"}, args.HasValue) {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "대화 조회는 페이지·budget 옵션을 지원하지 않습니다", "현재 서버에 보관된 대화만 한 번에 조회합니다"))
	}
	if len(args.Positionals) != maxPositionals || !args.HasValue("--work") || !args.HasValue("--workspace") || !userapi.ValidThreadReadID(args.Value("--work")) || !userapi.ValidThreadReadID(args.Workspace) || (maxPositionals == 1 && !userapi.ValidThreadReadID(args.Positionals[0])) {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "워크스페이스·작업 ID와 show의 대화 ID가 필요합니다", "riido thread --help"))
	}
	work, cliErr := userRelationID(args.Value("--work"))
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	var result any
	if raw[0] == "list" {
		result, err = session.Threads(ctx, args.Workspace, work)
	} else {
		result, err = session.Thread(ctx, args.Workspace, work, args.Positionals[0])
	}
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	env := contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "live": true, "history": result})
	env.Warnings = append(env.Warnings, contextcli.Warning{Code: "RETAINED_HISTORY", Message: "현재 권한으로 볼 수 있는 서버 보관분입니다. 전체 과거 대화·메시지의 완전성은 보장하지 않습니다"})
	return c.succeed(args.JSON, env, func(w io.Writer, _ any) { renderThreads(w, result) })
}

func printThreadUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido thread list --work <ID> --workspace <WORKSPACE-ID> [--json]")
	fmt.Fprintln(w, "  riido thread show <THREAD-ID> --work <ID> --workspace <WORKSPACE-ID> [--json]")
	fmt.Fprintln(w, "    User mode only; fresh task-scoped ticket, current visible server-retained history, not all-time history. No pagination. 16 MiB response limit; agent mode: exit 8.")
}

func renderThreads(w io.Writer, result any) {
	fmt.Fprintln(w, "현재 권한으로 볼 수 있는 서버 보관 대화입니다. 전체 과거 기록은 보장하지 않습니다.")
	switch value := result.(type) {
	case *userapi.ThreadList:
		for _, thread := range value.Threads {
			fmt.Fprintf(w, "  대화 %q (conversation=%q, agent=%q, run=%q, 상태=%q, 메시지=%d)\n", thread.ThreadID, thread.ConversationID, thread.AgentID, thread.RunID, thread.WorkStatus, thread.MessageCount)
		}
		if len(value.Threads) == 0 {
			fmt.Fprintln(w, "  현재 보관분에서 볼 수 있는 대화가 없습니다.")
		}
	case *userapi.ThreadShow:
		fmt.Fprintf(w, "대화 %q (conversation=%q, agent=%q, run=%q)\n", value.Thread.ThreadID, value.Thread.ConversationID, value.Thread.AgentID, value.Thread.RunID)
		for _, message := range value.Thread.Messages {
			fmt.Fprintf(w, "  %q %q: body=%q result=%q display=%q\n", message.Role, message.MessageID, message.Body, message.ResultMessage, message.DisplayMessage)
		}
	}
}
