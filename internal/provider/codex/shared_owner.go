package codex

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process/childenv"
	"github.com/teamswyg/riido-daemon/internal/process/execwindow"
)

var (
	errSharedOwnerBusy        = errors.New("codex app already has an active or pending turn")
	errSharedQueueConsumed    = errors.New("codex queued submission already consumed")
	errSharedOwnerConnect     = errors.New("codex_owner_connect_failed")
	errSharedOwnerInitialize  = errors.New("codex_owner_initialize_failed")
	errSharedOwnerRejected    = errors.New("codex_owner_start_rejected")
	errSharedOwnerUnconfirmed = errors.New("codex_owner_start_unconfirmed")
)

func sharedOwnerFailureMessage(err error) string {
	switch {
	case errors.Is(err, errSharedOwnerConnect):
		return "Codex의 기존 대화에 연결하지 못했습니다. 실행 중인 작업이 끝난 뒤, 에이전트가 실행되는 Mac에서 해당 대화를 소유한 Codex 앱 또는 CLI를 종료하고 Riido 답글로 다시 요청해주세요. 기존 대화에서 이어갑니다. [codex_owner_connect_failed]"
	case errors.Is(err, errSharedOwnerInitialize):
		return "Codex 연결 초기화에 실패했습니다. 에이전트가 실행되는 Mac의 Codex 앱과 CLI 버전을 확인해주세요. [codex_owner_initialize_failed]"
	case errors.Is(err, errSharedOwnerRejected):
		return "Codex가 기존 대화의 후속 실행을 거절했습니다. 해당 Codex 대화의 오류나 대기 중인 요청을 확인한 뒤 Riido 답글로 다시 요청해주세요. [codex_owner_start_rejected]"
	default:
		return "Codex 후속 실행의 시작 응답을 확인하지 못했습니다. 재요청 전에 해당 Codex 대화의 결과와 대기열을 확인해주세요. [codex_owner_start_unconfirmed]"
	}
}

type sharedOwnerClient struct {
	wire   *sharedOwnerWire
	close  func()
	nextID int64
}

func openSharedOwner(ctx context.Context, req agentbridge.StartRequest) (*sharedOwnerClient, error) {
	exe := req.Executable
	if exe == "" {
		exe = DefaultExecutable
	}
	// proxy only connects to the existing control socket. It neither starts a
	// daemon nor changes the owner's model, cwd, sandbox or approval policy.
	cmd := exec.CommandContext(ctx, exe, "app-server", "proxy")
	execwindow.Hide(cmd)
	cmd.Env = append(childenv.Environ(), childenv.Scrub(buildEnv(req.Env, StartOptions{}))...)
	cmd.Dir = req.Cwd
	return openSharedOwnerCommand(ctx, cmd)
}

func openSharedOwnerCommand(ctx context.Context, cmd *exec.Cmd) (*sharedOwnerClient, error) {
	cmd.WaitDelay = time.Second
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errSharedOwnerConnect
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, errSharedOwnerConnect
	}
	if err := cmd.Start(); err != nil {
		return nil, errSharedOwnerConnect
	}
	stop := context.AfterFunc(ctx, func() { _ = stdout.Close(); _ = stdin.Close() })
	closeProxy := func() {
		stop()
		_ = stdin.Close()
		_ = stdout.Close()
		// Only the short-lived relay is ours; never signal the owning app-server.
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}
	wire, err := upgradeSharedOwner(stdout, stdin)
	if err != nil {
		closeProxy()
		return nil, errSharedOwnerConnect
	}
	client := &sharedOwnerClient{wire: wire, close: closeProxy}
	_, err = client.request("initialize", map[string]any{
		"clientInfo":   map[string]any{"name": "riido-shared-writer", "version": "1"},
		"capabilities": map[string]any{"experimentalApi": true},
	})
	if err == nil {
		err = client.send(map[string]any{"method": "initialized"})
	}
	if err != nil {
		client.close()
		return nil, errSharedOwnerInitialize
	}
	return client, nil
}

func (c *sharedOwnerClient) send(frame map[string]any) error {
	body, err := json.Marshal(frame)
	if err != nil {
		return err
	}
	return c.wire.write(1, body)
}

func (c *sharedOwnerClient) request(method string, params map[string]any) (map[string]any, error) {
	c.nextID++
	if err := c.send(map[string]any{"id": c.nextID, "method": method, "params": params}); err != nil {
		return nil, err
	}
	for range 512 {
		body, err := c.wire.read()
		if err != nil {
			return nil, err
		}
		var frame map[string]any
		if err := json.Unmarshal(body, &frame); err != nil {
			return nil, err
		}
		if frame["method"] != nil {
			continue
		} // Notifications and server requests belong to the owner.
		if id, ok := rpcID(frame); !ok || id != c.nextID {
			continue
		}
		if frame["error"] != nil {
			// Do not surface raw provider errors or notifications (which can
			// contain private conversation data or credential-bearing paths).
			message := codexRPCErrorMessage(frame)
			if method == "thread/queue/start" {
				if message == "thread already has an active or pending turn" {
					return nil, errSharedOwnerBusy
				}
				if message == "queued submission not found: "+stringField(params, "queuedSubmissionId") {
					return nil, errSharedQueueConsumed
				}
			}
			return nil, errSharedOwnerRejected
		}
		result, ok := frame["result"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid Codex control response")
		}
		return result, nil
	}
	return nil, errors.New("codex control response count exceeded")
}

func startSharedOwner(ctx context.Context, req agentbridge.StartRequest, threadID, submissionID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	client, err := openSharedOwner(ctx, req)
	if err != nil {
		return err
	}
	defer client.close()
	result, err := client.request("thread/queue/start", map[string]any{"threadId": threadID, "queuedSubmissionId": submissionID})
	if err != nil {
		if errors.Is(err, errSharedOwnerBusy) || errors.Is(err, errSharedQueueConsumed) || errors.Is(err, errSharedOwnerRejected) {
			return err
		}
		return errSharedOwnerUnconfirmed
	}
	if stringField(mapField(result, "turn"), "id") == "" {
		return errSharedOwnerUnconfirmed
	}
	return nil
}
