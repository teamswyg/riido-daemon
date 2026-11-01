package main

import (
	"fmt"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

// Capture settings once. Workspace defaults apply only to component keys,
// preserving existing ID commands and their explicit workspace rejection.
func (c contextIO) parseUserWorkArgs(raw []string, spec contextcli.ArgSpec, component bool) (contextcli.Args, *contextcli.CLIError) {
	spec.UseWorkspace = false
	args, settings, err := c.parseSettings(raw, spec)
	if err != nil || settings == nil || !component || (args.HasValue("--kind") && args.Value("--kind") != "component") {
		return args, err
	}
	value := args.Value("--work")
	if len(args.Positionals) > 0 {
		value = args.Positionals[0]
	}
	ref, parseErr := contextcli.ParseWorkRef(value)
	if parseErr == nil && ref.Kind == contextcli.WorkRefKey {
		if err = args.ApplyDefault("--workspace", settings.resolved.Workspace); err != nil {
			return args, err
		}
	}
	return args, nil
}

func userWorkTarget(args contextcli.Args, raw string) (string, string, *contextcli.CLIError) {
	ref, err := contextcli.ParseWorkRef(raw)
	if err != nil {
		return "", "", err
	}
	team := strings.TrimSpace(args.Value("--team-id"))
	if team != "" {
		team, err = userRelationID(team)
		if err != nil {
			return "", "", err
		}
	} else if args.HasValue("--team-id") {
		return "", "", contextcli.Fail(contextcli.CodeUsage, "팀 ID가 비어 있습니다", "--team-id에 실제 팀 ID를 지정하세요")
	}
	if ref.Kind == contextcli.WorkRefKey {
		if !(userapi.WorkReferenceRequest{Reference: ref.Value, WorkspaceID: args.Workspace, TeamID: team}).Valid() || (args.HasValue("--workspace") && args.Workspace == "") || (team == "" && args.Workspace == "") {
			return "", "", contextcli.Fail(contextcli.CodeUsage, "표시 키에는 워크스페이스 또는 팀이 필요합니다", "--workspace <WORKSPACE-ID> 또는 --team-id <TEAM-ID>를 지정하세요")
		}
	} else {
		if team == "" {
			return "", "", contextcli.Fail(contextcli.CodeUsage, "작업 ID와 --team-id가 필요합니다", "작업의 실제 팀 ID를 지정하세요")
		}
		if args.HasValue("--workspace") {
			return "", "", contextcli.Fail(contextcli.CodeUnsupported, "ID 조회에는 --workspace를 지정할 수 없습니다", "작업 ID와 실제 팀 ID만 지정하세요")
		}
	}
	return ref.Value, team, nil
}

func userWorkBodyHint(work, team, workspace, version string) string {
	hint := fmt.Sprintf("riido work body %q", work)
	if team != "" {
		hint += fmt.Sprintf(" --team-id %q", team)
	}
	if workspace != "" {
		hint += fmt.Sprintf(" --workspace %q", workspace)
	}
	return hint + fmt.Sprintf(" --part metadata --version %q --offset 0 --limit 65536 --json", version)
}
