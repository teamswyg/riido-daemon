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
	Name       string               `json:"name"`
	Status     doctorStatus         `json:"status"`
	Detail     string               `json:"detail"`
	ErrorCode  contextcli.ErrorCode `json:"error_code,omitempty"`
	NextAction string               `json:"next_action,omitempty"`
}

type doctorData struct {
	Mode            contextcli.Mode `json:"mode"`
	Healthy         bool            `json:"healthy"`
	APIVerified     bool            `json:"api_verified"`
	Checks          []doctorCheck   `json:"checks"`
	APIOrigin       string          `json:"api_origin,omitempty"`
	APIOriginSource string          `json:"api_origin_source,omitempty"`
}

// runDoctor separates static installation checks from an explicitly requested
// authenticated user API read. Agent diagnosis never opens user storage.
func runDoctor(c contextIO, raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{Bools: []string{"--check-api"}})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printDoctorUsage(c.stderr)
		return nil
	}
	env := c.environment()
	if env.Mode == contextcli.ModeUser && args.Bool("--check-api") {
		if cliErr := checkUserOptions(args); cliErr != nil {
			return c.fail(args.JSON, requestID, cliErr)
		}
	}
	checks := []doctorCheck{c.doctorBinary(), c.doctorPath(), {Name: "mode", Status: doctorOK, Detail: string(env.Mode)}}
	checks = append(checks, c.doctorInstallation(env)...)
	daemonCheck, reachable := c.doctorDaemon(env)
	checks = append(checks, daemonCheck, c.doctorContextRead(env, requestID, reachable))
	data := doctorData{Mode: env.Mode, Healthy: true}
	var apiErr *contextcli.CLIError
	if env.Mode == contextcli.ModeUser {
		origin, source, originErr := c.userAPIOrigin()
		if originErr != nil {
			apiErr = userAuthError(originErr)
			checks = append(checks, doctorErrorCheck("api_origin", apiErr))
		} else {
			data.APIOrigin, data.APIOriginSource = origin, source
			checks = append(checks, doctorCheck{Name: "api_origin", Status: doctorOK, Detail: origin + " (" + source + ")"})
		}
	}
	checks = append(checks, doctorCheck{Name: "user_auth", Status: doctorSkip, Detail: "설치 진단은 자격을 읽지 않습니다. 인증은 명시적인 --check-api에서만 확인합니다"})
	apiCheck, verified, requestErr := c.doctorAPI(args, env, apiErr)
	checks = append(checks, apiCheck)
	if requestErr != nil {
		apiErr = requestErr
	}
	data.APIVerified = verified
	data.Checks = checks
	unsupported := env.Mode == contextcli.ModeAgent && args.Bool("--check-api")
	for _, check := range checks {
		if check.Status == doctorFail || check.Name == "helper_signature" && check.Status == doctorUnsupported {
			data.Healthy = false
		}
		if check.Name == "helper_signature" && check.Status == doctorUnsupported || env.Mode == contextcli.ModeAgent && check.Name == "context_read" && check.Status == doctorUnsupported {
			unsupported = true
		}
	}
	if unsupported {
		data.Healthy = false
	}
	if err := c.succeed(args.JSON, contextcli.NewEnvelope(requestID, data), renderDoctor); err != nil {
		return err
	}
	if apiErr != nil {
		return reportedExit{code: apiErr.ExitCode()}
	}
	if unsupported {
		return reportedExit{code: contextcli.ExitUnsupported}
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

func renderDoctor(w io.Writer, data any) {
	report, ok := data.(doctorData)
	if !ok {
		return
	}
	for _, check := range report.Checks {
		fmt.Fprintf(w, "[%s] %s: %s\n", check.Status, check.Name, check.Detail)
	}
}
