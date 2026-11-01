package main

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/teamswyg/riido-daemon/internal/cliconfig"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

var configKeys = []string{"profile", "api_origin", "auth_origin", "workspace", "team", "format", "timeout"}

func runConfig(c contextIO, raw []string) error {
	requestID := contextcli.NewRequestID()
	if len(raw) == 0 || isHelpArg(raw[0]) {
		printConfigUsage(c.stderr)
		if len(raw) == 0 {
			return c.fail(false, requestID, contextcli.Fail(contextcli.CodeUsage, "config 하위 명령이 필요합니다", "riido config --help"))
		}
		return nil
	}
	if !slices.Contains([]string{"list", "get", "set"}, raw[0]) {
		return c.fail(wantsJSON(raw), requestID, contextcli.Fail(contextcli.CodeUsage, "지원하지 않는 config 하위 명령입니다", "riido config --help"))
	}
	args, settings, cliErr := c.parseSettings(raw[1:], contextcli.ArgSpec{MaxPositionals: 2})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, contextcli.Fail(cliErr.Code, "config 입력이나 설정이 유효하지 않습니다", "riido config --help; 비밀값은 설정에 저장할 수 없습니다"))
	}
	if args.Help {
		printConfigUsage(c.stderr)
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "config는 사용자 모드에서만 지원합니다", "에이전트는 배정된 실행 범위를 사용하세요"))
	}
	if args.HasValue("--workspace") || args.HasValue("--cursor") {
		return c.fail(args.JSON, requestID, settingsError(cliconfig.ErrInvalid))
	}
	n := map[string]int{"list": 0, "get": 1, "set": 2}[raw[0]]
	if len(args.Positionals) != n || n > 0 && !slices.Contains(configKeys, args.Positionals[0]) {
		return c.fail(args.JSON, requestID, settingsError(cliconfig.ErrInvalid))
	}
	key := ""
	if n > 0 {
		key = args.Positionals[0]
	}
	if raw[0] == "set" {
		if key == "profile" && args.HasValue("--profile") {
			return c.fail(args.JSON, requestID, settingsError(cliconfig.ErrInvalid))
		}
		value := args.Positionals[1]
		if key == "workspace" && value != "" {
			return c.selectWorkspace(requestID, args, settings, value)
		}
		ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
		defer cancel()
		if err := settings.store.Set(ctx, settings.resolved, key, value); err != nil {
			return c.fail(args.JSON, requestID, settingsError(err))
		}
		profile := settings.resolved.Profile
		if key == "profile" {
			profile = value
		}
		return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, map[string]any{"saved": true, "profile": profile, "key": key}), func(w io.Writer, _ any) { fmt.Fprintf(w, "설정을 저장했습니다: %s\n", key) })
	}
	r := settings.resolved
	values := map[string]string{"profile": r.Profile, "api_origin": r.APIOrigin, "auth_origin": r.AuthOrigin, "workspace": r.Workspace, "team": r.Team, "format": r.Format, "timeout": r.Timeout}
	if key != "" {
		values = map[string]string{key: values[key]}
	}
	sources := map[string]string{}
	for name := range values {
		sources[name] = r.Sources[name]
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, map[string]any{"profile": r.Profile, "values": values, "sources": sources}), func(w io.Writer, _ any) {
		for _, name := range configKeys {
			if value, ok := values[name]; ok {
				fmt.Fprintf(w, "%s=%q (%s)\n", name, value, sources[name])
			}
		}
	})
}

func printConfigUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido config list [--profile PROFILE] [--json]")
	fmt.Fprintln(w, "  riido config get <KEY> [--profile PROFILE] [--json]")
	fmt.Fprintln(w, "  riido config set <KEY> <VALUE> [--profile PROFILE] [--json]")
	fmt.Fprintln(w, "    Keys: profile, api_origin, auth_origin, workspace, team, format, timeout. Profiles: local, staging, production.")
	fmt.Fprintln(w, "    User mode only; nonsecret preferences. Empty VALUE clears a profile setting. workspace is verified through the current login. local requires a literal loopback api_origin.")
}
