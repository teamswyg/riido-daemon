package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func runResource(c contextIO, raw []string) error {
	requestID := contextcli.NewRequestID()
	if len(raw) == 0 || isHelpArg(raw[0]) {
		printResourceUsage(c.stderr)
		if len(raw) == 0 {
			return c.fail(false, requestID, contextcli.Fail(contextcli.CodeUsage, "resource 하위 명령이 필요합니다", "riido resource --help"))
		}
		return nil
	}
	if raw[0] != "list" {
		return c.fail(wantsJSON(raw), requestID, contextcli.Fail(contextcli.CodeUsage, "지원하지 않는 resource 하위 명령입니다", "riido resource list --help"))
	}
	args, cliErr := contextcli.ParseArgs(raw[1:], contextcli.ArgSpec{Values: []string{"--work", "--team-id", "--page", "--budget"}})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printResourceUsage(c.stderr)
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "resource list는 아직 사용자 모드에서만 지원합니다", "에이전트는 riido context get으로 허용된 자료를 조회하세요"))
	}
	if !args.HasValue("--work") || !args.HasValue("--team-id") {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "--work와 --team-id가 필요합니다", "riido resource list --work <ID> --team-id <TEAM-ID>"))
	}
	if slices.ContainsFunc([]string{"--workspace", "--cursor", "--page", "--budget"}, args.HasValue) {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "자료 목록은 범위·페이지 옵션을 지원하지 않습니다", "작업 ID와 팀 ID만 지정하세요. 큰 메타데이터는 같은 버전으로 자동 이어 읽습니다"))
	}
	work, cliErr := userRelationID(args.Value("--work"))
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
	resources, err := session.Resources(ctx, work, team)
	var limitErr *userapi.ResourceMetadataLimitError
	if errors.As(err, &limitErr) {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeIncomplete, "메타데이터가 자동 목록 조립 한도(16 MiB)를 넘었습니다", fmt.Sprintf("riido work body %q --team-id %q --part metadata --version %q --offset 0 --limit 65536 --json", work, team, limitErr.Version)))
	}
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	env := contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "live": true, "resources": resources})
	if !resources.Complete {
		env.Warnings = append(env.Warnings, contextcli.Warning{Code: "RESOURCES_INCOMPLETE", Message: "일부 자료의 조회를 완료하지 못했습니다. partsStatus를 확인하세요"})
	}
	return c.succeed(args.JSON, env, func(w io.Writer, _ any) { renderResources(w, resources) })
}

func printResourceUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido resource list --work <ID> --team-id <TEAM-ID> [--json]   (user stored links, attachments and Figma; no external content fetch)")
	fmt.Fprintln(w, "    User mode only; original-version metadata paging is automatic up to 16 MiB, then use work body --part metadata. Agent mode: unsupported, exit 8.")
}

func renderResources(w io.Writer, r *userapi.Resources) {
	value := func(s *string) string {
		if s == nil {
			return "(없음)"
		}
		return *s
	}
	fmt.Fprintf(w, "작업 %q의 자료\n", r.WorkID)
	for _, link := range r.Links {
		fmt.Fprintf(w, "  링크 %q: %q (id=%q)\n", value(link.Title), value(link.URL), link.ID)
	}
	for _, file := range r.Attachments {
		fmt.Fprintf(w, "  첨부 %q: %q (id=%q)\n", value(file.FileName), value(file.FileLocation), file.ID)
	}
	if r.Figma != nil {
		fmt.Fprintf(w, "  Figma: %q (id=%q)\n", value(r.Figma.URL), r.Figma.ID)
	}
	if r.Complete && len(r.Links) == 0 && len(r.Attachments) == 0 && r.Figma == nil {
		fmt.Fprintln(w, "  연결된 자료가 없습니다.")
	}
}
