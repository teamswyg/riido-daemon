package main

import (
	"errors"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

// Workspace defaults apply to both forms; team defaults apply only to keys.
// Parsing settings once preserves the distinction between explicit and stored
// team selectors so existing ID calls do not acquire a new constraint.
func (c contextIO) parseCPWorkArgs(raw []string, spec contextcli.ArgSpec) (contextcli.Args, *contextcli.CLIError) {
	spec.Values = append(spec.Values, "--team-id")
	args, settings, err := c.parseSettingsWithTeamDefault(raw, spec, false)
	if err != nil || settings == nil {
		return args, err
	}
	ref, parseErr := contextcli.ParseWorkRef(args.Value("--work"))
	if parseErr == nil && ref.Kind == contextcli.WorkRefKey {
		if err = args.ApplyDefault("--team-id", settings.resolved.Team); err != nil {
			return args, err
		}
	}
	return args, nil
}

func userCPWorkTarget(args contextcli.Args) (string, string, *contextcli.CLIError) {
	ref, err := contextcli.ParseWorkRef(args.Value("--work"))
	if err != nil {
		return "", "", err
	}
	if ref.Kind == contextcli.WorkRefKey {
		return userWorkTarget(args, ref.Value)
	}
	if args.HasValue("--team-id") || args.HasValue("--team") || !userapi.ValidThreadReadID(args.Workspace) {
		return "", "", contextcli.Fail(contextcli.CodeUsage, "작업 ID 조회에는 워크스페이스가 필요하며 팀 옵션은 지원하지 않습니다", "--workspace <WORKSPACE-ID>를 지정하세요")
	}
	return ref.Value, "", nil
}

func userCPReadError(err error) *contextcli.CLIError {
	if errors.Is(err, userapi.ErrAmbiguousWorkReference) {
		return contextcli.Fail(contextcli.CodeConflict, "표시 키에 해당하는 작업이 여러 개입니다", "표시 키와 --team-id로 범위를 좁히거나 --work <ID> --workspace <WORKSPACE-ID>를 지정하세요")
	}
	return userAuthError(err)
}
