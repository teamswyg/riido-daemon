package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/selfupdate"
)

// updateTimeout bounds the package download and checks of one update.
const updateTimeout = 5 * time.Minute

type updateData struct {
	Channel        string `json:"channel"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	UpdateNeeded   bool   `json:"update_needed"`
	Updated        bool   `json:"updated"`
	InstallDir     string `json:"install_dir"`
}

// runUpdate (RIID-6021) replaces an installer-managed CLI with the newest
// package of its channel, using the installer's CDN assets and checks.
func runUpdate(c contextIO, raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{Bools: []string{"--check"}, Values: []string{"--channel"}})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printUpdateUsage(c.stderr)
		return nil
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnsupported, "riido update는 사용자 터미널에서만 실행할 수 있습니다", "사용자가 직접 riido update 를 실행하세요"))
	}
	current := strings.TrimSpace(binaryVersion)
	channel := selfupdate.ChannelForVersion(current)
	if args.HasValue("--channel") {
		channel = args.Value("--channel")
		if !selfupdate.ValidChannel(channel) {
			return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage, "--channel은 staging 또는 latest만 받습니다", "riido update --channel staging"))
		}
	}
	root, cliErr := c.updateRoot(current)
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	verify := c.verifyHelperSignature
	if verify == nil {
		verify = securecredential.VerifyPackagedHelper
	}
	updater := selfupdate.Updater{BaseURL: selfupdate.DefaultBaseURL, Client: &http.Client{Timeout: updateTimeout}, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, Verify: verify}
	if c.updateBaseURL != "" {
		updater.BaseURL = c.updateBaseURL
	}
	if c.updateClient != nil {
		updater.Client = c.updateClient
	}
	if c.updateGOOS != "" {
		updater.GOOS, updater.GOARCH = c.updateGOOS, c.updateGOARCH
	}
	ctx, cancel := context.WithTimeout(context.Background(), updateTimeout)
	defer cancel()
	staged, err := updater.Stage(ctx, channel, root)
	if err != nil {
		return c.fail(args.JSON, requestID, updateError(err))
	}
	defer staged.Cleanup()
	data := updateData{Channel: channel, CurrentVersion: current, LatestVersion: staged.Version, UpdateNeeded: staged.Version != current, InstallDir: root}
	if data.UpdateNeeded && !args.Bool("--check") {
		if err = selfupdate.Install(staged, root); err != nil {
			return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUnavailable, "새 버전으로 바꾸지 못했습니다", "riido update 를 다시 실행하거나 설치 스크립트로 다시 설치하세요"))
		}
		data.Updated = true
	}
	env := contextcli.NewEnvelope(requestID, map[string]any{"mode": "user", "update": data})
	return c.succeed(args.JSON, env, func(w io.Writer, _ any) {
		switch {
		case !data.UpdateNeeded:
			fmt.Fprintf(w, "이미 최신 버전입니다: %s (%s)\n", data.CurrentVersion, data.Channel)
		case data.Updated:
			fmt.Fprintf(w, "✓ %s → %s 로 업데이트했습니다 (%s)\n", data.CurrentVersion, data.LatestVersion, data.Channel)
			fmt.Fprintln(w, "  riido --version 으로 확인하세요. 첫 조회에서 Touch ID가 한 번 뜰 수 있어요.")
		default:
			fmt.Fprintf(w, "새 버전이 있습니다: %s → %s (%s)\n", data.CurrentVersion, data.LatestVersion, data.Channel)
			fmt.Fprintln(w, "  riido update 로 업데이트하세요.")
		}
	})
}

// updateRoot accepts only the installer layout: the resolved riido binary with
// an adjacent VERSION equal to the running version, outside any app bundle.
func (c contextIO) updateRoot(current string) (string, *contextcli.CLIError) {
	unsupported := func(message string) (string, *contextcli.CLIError) {
		return "", contextcli.Fail(contextcli.CodeUnsupported, message, "설치 스크립트로 다시 설치하세요: curl -fsSL https://raw.githubusercontent.com/teamswyg/riido-daemon/main/scripts/install-riido-daemon.sh | sh")
	}
	exe, err := c.executable()
	if err != nil {
		return unsupported("실행 파일 위치를 확인하지 못했습니다")
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return unsupported("실행 파일 위치를 확인하지 못했습니다")
	}
	if strings.Contains(filepath.ToSlash(resolved), ".app/Contents/") {
		return "", contextcli.Fail(contextcli.CodeUnsupported, "Riido 앱에 포함된 CLI는 앱과 함께 업데이트됩니다", "Riido 앱을 업데이트하세요")
	}
	if filepath.Base(resolved) != "riido" {
		return unsupported("설치 스크립트로 설치한 riido가 아닙니다")
	}
	root := filepath.Dir(resolved)
	raw, err := doctorReadRegular(filepath.Join(root, "VERSION"), 4096)
	if err != nil || current == "" || current == "dev" || strings.TrimSpace(string(raw)) != current {
		return unsupported("설치 정보(VERSION)가 실행 중인 버전과 맞지 않아 자동 업데이트할 수 없습니다")
	}
	return root, nil
}

func updateError(err error) *contextcli.CLIError {
	switch {
	case errors.Is(err, selfupdate.ErrUnsupported):
		return contextcli.Fail(contextcli.CodeUnsupported, "이 OS·아키텍처는 riido update를 지원하지 않습니다", "설치 스크립트로 다시 설치하세요")
	case errors.Is(err, selfupdate.ErrInvalid):
		return contextcli.Fail(contextcli.CodeUnavailable, "내려받은 패키지 검증에 실패해 설치하지 않았습니다", "잠시 뒤 다시 실행하세요. 계속되면 riido doctor 결과와 함께 알려 주세요")
	case errors.Is(err, os.ErrPermission):
		return contextcli.Fail(contextcli.CodeUnavailable, "설치 폴더에 쓸 권한이 없습니다", "설치한 사용자로 다시 실행하세요")
	case errors.Is(err, context.DeadlineExceeded):
		return contextcli.Fail(contextcli.CodeUnavailable, "업데이트 다운로드 시간이 초과됐습니다", "네트워크를 확인하고 다시 실행하세요")
	default:
		return contextcli.Fail(contextcli.CodeUnavailable, "업데이트를 내려받지 못했습니다", "네트워크를 확인하고 다시 실행하세요")
	}
}

func printUpdateUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido update [--check] [--channel staging|latest] [--json]   (installer-managed CLI only; verifies SHA256SUMS and the helper signature before replacing)")
}
