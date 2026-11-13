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
	if err != nil {
		return args, err
	}
	version, versionErr := userLocatorVersion(args)
	if versionErr != nil {
		return args, versionErr
	}
	if settings == nil || !component || (version != 2 && args.HasValue("--kind") && args.Value("--kind") != "component") {
		return args, err
	}
	value := args.Value("--work")
	if len(args.Positionals) > 0 {
		value = args.Positionals[0]
	}
	ref, parseErr := contextcli.ParseWorkRef(value)
	if version == 2 || parseErr == nil && ref.Kind == contextcli.WorkRefKey {
		if err = args.ApplyDefault("--workspace", settings.resolved.Workspace); err != nil {
			return args, err
		}
	}
	return args, nil
}

func userWorkTarget(args contextcli.Args, raw string) (string, string, *contextcli.CLIError) {
	version, versionErr := userLocatorVersion(args)
	if versionErr != nil {
		return "", "", versionErr
	}
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
	if version == 2 {
		if !(userapi.WorkReferenceRequest{Reference: ref.Value, WorkspaceID: args.Workspace, TeamID: team, LocatorVersion: 2}).Valid() {
			return "", "", contextcli.Fail(contextcli.CodeUsage, "버전 2 작업 조회에는 워크스페이스와 작업 키 또는 ID가 필요합니다", "--locator-version 2 --workspace <WORKSPACE-ID>를 지정하세요")
		}
	} else if ref.Kind == contextcli.WorkRefKey {
		if !(userapi.WorkReferenceRequest{Reference: ref.Value, WorkspaceID: args.Workspace, TeamID: team}).Valid() || (args.HasValue("--workspace") && args.Workspace == "") || (team == "" && args.Workspace == "") {
			return "", "", contextcli.Fail(contextcli.CodeUsage, "표시 키에는 워크스페이스 또는 팀이 필요합니다", "--workspace <WORKSPACE-ID> 또는 --team-id <TEAM-ID>를 지정하세요 (riido workspace use <WORKSPACE-ID>로 기본값을 저장할 수 있어요)")
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

func userLocatorVersion(args contextcli.Args) (int, *contextcli.CLIError) {
	if !args.HasValue("--locator-version") || args.Value("--locator-version") == "1" {
		return 1, nil
	}
	if args.Value("--locator-version") == "2" {
		return 2, nil
	}
	return 0, contextcli.Fail(contextcli.CodeUsage, "locator-version은 1 또는 2여야 합니다", "기존 API는 1, 워크스페이스 작업 조회는 2를 사용하세요")
}

func userV2Reference(args contextcli.Args, id, team string) userapi.WorkReferenceRequest {
	return userapi.WorkReferenceRequest{Reference: id, WorkspaceID: args.Workspace, TeamID: team, LocatorVersion: 2}
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
