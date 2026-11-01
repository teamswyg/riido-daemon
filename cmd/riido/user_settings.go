package main

import (
	"context"
	"errors"
	"slices"

	"github.com/teamswyg/riido-daemon/internal/cliconfig"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

type commandSettings struct {
	resolved cliconfig.Resolved
	store    cliconfig.Store
}

func (c contextIO) parseArgs(raw []string, spec contextcli.ArgSpec) (contextcli.Args, *contextcli.CLIError) {
	args, _, err := c.parseSettings(raw, spec)
	return args, err
}

// Capture preferences and origin once, before opening credentials. Agent calls
// retain the original parser and never read user preferences.
func (c contextIO) parseSettings(raw []string, spec contextcli.ArgSpec) (contextcli.Args, *commandSettings, *contextcli.CLIError) {
	userMode := c.environment().Mode != contextcli.ModeAgent
	spec.AllowProfile = userMode
	spec.AllowTeam = userMode && slices.Contains(spec.Values, "--team-id")
	args, cliErr := contextcli.ParseArgs(raw, spec)
	if cliErr != nil || args.Help || !userMode {
		return args, nil, cliErr
	}
	settings := &commandSettings{}
	if c.configDir != nil {
		dir, err := c.configDir()
		if err != nil {
			return args, settings, settingsError(cliconfig.ErrUnavailable)
		}
		settings.store.Dir = dir
	}
	config, err := settings.store.Load()
	if err != nil {
		return args, settings, settingsError(err)
	}
	getenv := func(key string) string {
		// A lower-priority environment value cannot invalidate an explicit flag.
		for env, flag := range map[string]string{"RIIDO_CLI_WORKSPACE": "--workspace", "RIIDO_CLI_TEAM": "--team-id", "RIIDO_CLI_FORMAT": "--format", "RIIDO_CLI_TIMEOUT": "--timeout"} {
			if key == env && (args.HasValue(flag) || env == "RIIDO_CLI_TEAM" && args.HasValue("--team") || env == "RIIDO_CLI_FORMAT" && args.Bool("--json")) {
				return ""
			}
		}
		return c.getenv(key)
	}
	resolved, err := cliconfig.Resolve(config, args.Value("--profile"), args.HasValue("--profile"), getenv)
	if err != nil {
		return args, settings, settingsError(err)
	}
	settings.resolved = resolved
	args.SettingsResolved, args.Profile, args.APIOrigin, args.APIOriginSource = true, resolved.Profile, resolved.APIOrigin, resolved.Sources["api_origin"]
	if args.HasValue("--team") && args.HasValue("--team-id") && args.Value("--team") != args.Value("--team-id") {
		return args, settings, contextcli.Fail(contextcli.CodeUsage, "--team과 --team-id가 서로 다릅니다", "한 가지 팀 ID를 사용하세요")
	}
	defaults := map[string]string{"--timeout": resolved.Timeout}
	if !args.Bool("--json") {
		defaults["--format"] = resolved.Format
	}
	if spec.UseWorkspace {
		defaults["--workspace"] = resolved.Workspace
	}
	if spec.AllowTeam {
		team := resolved.Team
		if args.HasValue("--team") {
			team = args.Value("--team")
		}
		// An explicitly empty alias must not silently reuse a stored team.
		if args.HasValue("--team") && team == "" {
			return args, settings, contextcli.Fail(contextcli.CodeUsage, "--team에 팀 ID가 필요합니다", "팀 ID를 지정하세요")
		}
		defaults["--team-id"] = team
	}
	for name, value := range defaults {
		if cliErr := args.ApplyDefault(name, value); cliErr != nil {
			return args, settings, cliErr
		}
	}
	return args, settings, nil
}

func settingsError(err error) *contextcli.CLIError {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return userAuthError(err)
	}
	if errors.Is(err, cliconfig.ErrInvalid) {
		return contextcli.Fail(contextcli.CodeUsage, "CLI 설정 형식이나 값이 유효하지 않습니다", "프로필·환경변수와 config.json을 확인하세요. 설정을 자동 초기화하지 않습니다")
	}
	if errors.Is(err, cliconfig.ErrConflict) {
		return contextcli.Fail(contextcli.CodeConflict, "명령 실행 중 CLI 설정이 변경됐습니다", "현재 설정을 확인하고 다시 실행하세요")
	}
	return contextcli.Fail(contextcli.CodeUnavailable, "CLI 설정을 읽거나 저장할 수 없습니다", "사용자 설정 디렉터리와 파일 권한을 확인하세요")
}
