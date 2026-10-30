package main

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (c contextIO) runUserTree(raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{MaxPositionals: 1, Values: []string{"--team-id", "--page", "--budget"}})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printWorkUsage(c.stderr)
		return nil
	}
	if len(args.Positionals) != 1 || !args.HasValue("--team-id") {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "작업 ID와 --team-id가 필요합니다", "riido work tree <ID> --team-id <TEAM-ID>"))
	}
	if slices.ContainsFunc([]string{"--workspace", "--cursor", "--page", "--budget"}, args.HasValue) {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "작업 계층 조회는 범위·페이지 옵션을 지원하지 않습니다", "작업 ID와 --team-id만 지정하세요. 하위 작업은 서버가 최대 10개 선택합니다"))
	}
	work, cliErr := userRelationID(args.Positionals[0])
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	team, cliErr := userRelationID(args.Value("--team-id"))
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	session, err := c.openUserSession(args)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	tree, err := session.Tree(ctx, work, team)
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	env := contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "live": true, "tree": tree})
	if tree.ContextStatus == "partial" {
		env.Warnings = append(env.Warnings, contextcli.Warning{Code: "CONTEXT_INCOMPLETE", Message: "서버 맥락의 일부 자료가 미완료입니다. 계층은 허용된 선택 범위만 표시합니다"})
	}
	return c.succeed(args.JSON, env, func(w io.Writer, _ any) { renderUserTree(w, tree) })
}

func renderUserTree(w io.Writer, tree *userapi.WorkTree) {
	// Quote all remote text, including IDs, to keep terminal control bytes inert.
	fmt.Fprintf(w, "작업: %q (id=%q, team=%q)\n", tree.Root.Title, tree.Root.ID, tree.Root.TeamID)
	labels := map[string]string{"project": "프로젝트", "milestone": "목표", "parentTask": "상위 작업"}
	for _, a := range tree.Ancestors {
		fmt.Fprintf(w, "  %s: %q (id=%q, team=%q)\n", labels[a.Relation], a.Work.Title, a.Work.ID, a.Work.TeamID)
	}
	fmt.Fprintf(w, "선택된 하위 작업: %d개 (최대 %d개, 한 단계)\n", len(tree.Children), tree.ChildrenLimit)
	for _, child := range tree.Children {
		fmt.Fprintf(w, "  - %q (id=%q, team=%q)\n", child.Title, child.ID, child.TeamID)
	}
	fmt.Fprintln(w, "현재 권한과 서버 선택 범위 안의 결과입니다.")
}
