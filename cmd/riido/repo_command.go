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

func runRepo(c contextIO, raw []string) error {
	requestID := contextcli.NewRequestID()
	if len(raw) == 0 || isHelpArg(raw[0]) {
		printRepoUsage(c.stderr)
		if len(raw) == 0 {
			return c.fail(false, requestID, contextcli.Fail(contextcli.CodeUsage, "repo 하위 명령이 필요합니다", "riido repo --help"))
		}
		return nil
	}
	if raw[0] != "list" {
		return c.fail(wantsJSON(raw), requestID, contextcli.Fail(contextcli.CodeUsage, "지원하지 않는 repo 하위 명령입니다", "riido repo list --help"))
	}
	args, cliErr := contextcli.ParseArgs(raw[1:], contextcli.ArgSpec{Values: []string{"--work", "--team-id", "--page", "--budget"}})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printRepoUsage(c.stderr)
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "repo list는 아직 사용자 모드에서만 지원합니다", "에이전트는 riido context get으로 허용된 자료를 조회하세요"))
	}
	if !args.HasValue("--work") || !args.HasValue("--team-id") {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "--work와 --team-id가 필요합니다", "riido repo list --work <ID> --team-id <TEAM-ID>"))
	}
	if slices.ContainsFunc([]string{"--workspace", "--cursor", "--page", "--budget"}, args.HasValue) {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "저장소 목록은 범위·페이지 옵션을 지원하지 않습니다", "작업 ID와 팀 ID만 지정하세요. 큰 메타데이터는 같은 버전으로 자동 이어 읽습니다"))
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
	repos, err := session.Repositories(ctx, work, team)
	var limitErr *userapi.ResourceMetadataLimitError
	if errors.As(err, &limitErr) {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeIncomplete, "메타데이터가 자동 목록 조립 한도(16 MiB)를 넘었습니다", fmt.Sprintf("riido work body %q --team-id %q --part metadata --version %q --offset 0 --limit 65536 --json", work, team, limitErr.Version)))
	}
	if err != nil {
		return c.fail(args.JSON, requestID, userAuthError(err))
	}
	env := contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "live": true, "repo": repos, "local_repository_authentication": map[string]any{"used": false, "availability": "not_checked"}})
	if !repos.Complete {
		env.Warnings = append(env.Warnings, contextcli.Warning{Code: "REPOSITORIES_INCOMPLETE", Message: "일부 저장소·PR의 조회를 완료하지 못했습니다. partsStatus를 확인하세요"})
	}
	return c.succeed(args.JSON, env, func(w io.Writer, _ any) { renderRepositories(w, repos) })
}

func printRepoUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido repo list --work <ID> --team-id <TEAM-ID> [--json]   (user stored repositories and PRs; local Git/GitHub/SSH authentication is not used or checked)")
	fmt.Fprintln(w, "    User mode only; original-version metadata paging is automatic up to 16 MiB, then use work body --part metadata. Agent mode: unsupported, exit 8.")
}

func renderRepositories(w io.Writer, r *userapi.RepositoryList) {
	fmt.Fprintf(w, "작업 %q의 저장소·PR\n", r.WorkID)
	fmt.Fprintln(w, "  이 명령에서는 로컬 Git/GitHub/SSH 인증을 사용하지 않습니다. 사용 가능 여부는 확인하지 않았습니다.")
	for _, repo := range r.Repositories {
		fmt.Fprintf(w, "  저장소 %q: %q (id=%q, private=%t, source=%q)\n", *repo.FullName, *repo.RepositoryURL, repo.ID, *repo.IsPrivate, *repo.Source)
	}
	for _, pr := range r.PullRequests {
		fmt.Fprintf(w, "  PR #%d %q: %q (id=%q, state=%q, comments=%d)\n", *pr.Number, *pr.Title, *pr.HTMLURL, pr.ID, *pr.State, *pr.CommentsCount)
	}
	if r.Complete && len(r.Repositories) == 0 && len(r.PullRequests) == 0 {
		fmt.Fprintln(w, "  연결된 저장소·PR이 없습니다.")
	}
}
