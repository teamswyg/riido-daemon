package main

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

type doctorStatus string

const (
	doctorOK          doctorStatus = "ok"
	doctorWarn        doctorStatus = "warn"
	doctorFail        doctorStatus = "fail"
	doctorSkip        doctorStatus = "skip"
	doctorUnsupported doctorStatus = "unsupported"
)

const doctorProbeTimeout = 3 * time.Second

type doctorCheck struct {
	Name   string       `json:"name"`
	Status doctorStatus `json:"status"`
	Detail string       `json:"detail"`
}

type doctorData struct {
	Mode    contextcli.Mode `json:"mode"`
	Healthy bool            `json:"healthy"`
	Checks  []doctorCheck   `json:"checks"`
}

// runDoctor diagnoses the local setup. A stored credential file alone never
// counts as a working login; remote checks run only with --check-api.
func runDoctor(c contextIO, raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{Bools: []string{"--check-api"}})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		fmt.Fprintln(c.stderr, "  riido doctor [--check-api] [--json]")
		return nil
	}
	env := c.environment()
	checks := []doctorCheck{c.doctorBinary(), c.doctorPath(), {Name: "mode", Status: doctorOK, Detail: string(env.Mode)}}
	daemonCheck, reachable := c.doctorDaemon(env)
	checks = append(checks, daemonCheck, c.doctorContextRead(env, requestID, reachable),
		doctorCheck{Name: "user_auth", Status: doctorUnsupported, Detail: "사용자 모드 로그인은 아직 지원하지 않습니다"},
		doctorAPI(args.Bool("--check-api")))
	data := doctorData{Mode: env.Mode, Healthy: true, Checks: checks}
	for _, check := range checks {
		if check.Status == doctorFail {
			data.Healthy = false
		}
	}
	if err := c.succeed(args.JSON, contextcli.NewEnvelope(requestID, data), renderDoctor); err != nil {
		return err
	}
	if !data.Healthy {
		return reportedExit{code: contextcli.ExitUnavailable}
	}
	return nil
}

func (c contextIO) doctorBinary() doctorCheck {
	exe, err := c.executable()
	if err != nil {
		return doctorCheck{Name: "binary", Status: doctorWarn, Detail: "실행 파일 경로를 확인하지 못했습니다"}
	}
	return doctorCheck{Name: "binary", Status: doctorOK, Detail: exe + " (" + c.cliVersion() + ")"}
}

func (c contextIO) doctorPath() doctorCheck {
	found, err := c.lookPath("riido")
	if err != nil {
		return doctorCheck{Name: "path", Status: doctorWarn, Detail: "터미널 PATH에서 riido를 찾지 못했습니다. Riido Desktop 설정에서 터미널 연결을 켜세요"}
	}
	exe, err := c.executable()
	if err == nil && sameFile(found, exe) {
		return doctorCheck{Name: "path", Status: doctorOK, Detail: found}
	}
	return doctorCheck{Name: "path", Status: doctorWarn, Detail: "PATH의 riido가 지금 실행한 바이너리와 다릅니다: " + found}
}

func sameFile(a, b string) bool {
	ra, errA := filepath.EvalSymlinks(a)
	rb, errB := filepath.EvalSymlinks(b)
	if errA != nil || errB != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return ra == rb
}

func (c contextIO) doctorDaemon(env contextcli.Environment) (doctorCheck, bool) {
	missing := doctorWarn
	if env.Mode == contextcli.ModeAgent {
		missing = doctorFail
	}
	socket, err := c.daemonSocket(env)
	if err != nil {
		return doctorCheck{Name: "daemon_ipc", Status: missing, Detail: "daemon 소켓 경로를 찾지 못했습니다"}, false
	}
	body, cliErr := c.daemonRoundTrip(socket, doctorProbeTimeout, daemonRequest{Method: daemonMethodStatus})
	if cliErr != nil {
		return doctorCheck{Name: "daemon_ipc", Status: missing, Detail: cliErr.Message + ": " + socket}, false
	}
	var status struct {
		DaemonVersion string `json:"daemon_version"`
	}
	if err := json.Unmarshal(body, &status); err != nil || status.DaemonVersion == "" {
		return doctorCheck{Name: "daemon_ipc", Status: doctorWarn, Detail: "daemon이 응답했지만 버전을 읽지 못했습니다: " + socket}, true
	}
	return doctorCheck{Name: "daemon_ipc", Status: doctorOK, Detail: status.DaemonVersion + " @ " + socket}, true
}

func (c contextIO) doctorContextRead(env contextcli.Environment, requestID string, reachable bool) doctorCheck {
	if !reachable {
		return doctorCheck{Name: "context_read", Status: doctorSkip, Detail: "daemon에 연결되지 않아 확인하지 않았습니다"}
	}
	socket, err := c.daemonSocket(env)
	if err != nil {
		return doctorCheck{Name: "context_read", Status: doctorSkip, Detail: "daemon 소켓 경로를 찾지 못했습니다"}
	}
	body, cliErr := c.daemonRoundTrip(socket, doctorProbeTimeout, contextcli.NewContextReadRequest(env, requestID, "capabilities", nil))
	if cliErr != nil {
		return doctorCheck{Name: "context_read", Status: doctorWarn, Detail: cliErr.Message}
	}
	if _, replyErr := contextcli.DecodeContextReadReply(body); replyErr != nil && replyErr.Code == contextcli.CodeContextReadUnsupported {
		return doctorCheck{Name: "context_read", Status: doctorUnsupported, Detail: replyErr.Message}
	}
	return doctorCheck{Name: "context_read", Status: doctorOK, Detail: "daemon이 context_read를 지원합니다"}
}

func doctorAPI(requested bool) doctorCheck {
	if requested {
		return doctorCheck{Name: "api", Status: doctorSkip, Detail: "서버 인증 계약 확정 전이라 원격 점검을 하지 않습니다"}
	}
	return doctorCheck{Name: "api", Status: doctorSkip, Detail: "--check-api 를 주면 서버 도달을 점검합니다"}
}

func renderDoctor(w io.Writer, data any) {
	report, ok := data.(doctorData)
	if !ok {
		return
	}
	for _, check := range report.Checks {
		fmt.Fprintf(w, "[%s] %s: %s\n", check.Status, check.Name, check.Detail)
	}
}
