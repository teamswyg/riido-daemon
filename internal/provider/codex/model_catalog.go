package codex

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

type ModelInfo struct {
	ModelID   string
	Label     string
	IsDefault bool
}

// ModelCatalog reads the native catalog without creating a thread or running
// inference. Configured model names are not evidence of account availability.
func ModelCatalog(ctx context.Context, env agentbridge.DetectEnv) ([]ModelInfo, error) {
	exe, ok := detectutil.ResolveExecutable(DefaultExecutable, envValue(env, EnvOverride))
	if !ok {
		return nil, errors.New("codex model/list: executable unavailable")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	spawn, err := BuildStart(agentbridge.StartRequest{Executable: exe, Env: detectutil.EnvMapWithLaunchPATH(nil)}, StartOptions{})
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, spawn.Executable, spawn.Args...)
	cmd.WaitDelay = time.Second
	// Preserve credential/config variables while using the discovered PATH.
	cmd.Env = append(cmd.Environ(), spawn.Env...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	stopReader := context.AfterFunc(ctx, func() { _ = stdout.Close() })
	defer stopReader()
	defer func() { cancel(); _ = stdin.Close(); _ = cmd.Wait() }()
	encoder := json.NewEncoder(stdin)
	decoder := json.NewDecoder(io.LimitReader(stdout, 2*1024*1024))
	request := func(id int, method string, params map[string]any) (map[string]any, error) {
		if err := encoder.Encode(map[string]any{"id": id, "method": method, "params": params}); err != nil {
			return nil, err
		}
		for {
			var frame map[string]any
			if err := decoder.Decode(&frame); err != nil {
				return nil, errors.New("codex model/list: probe did not complete")
			}
			responseID, ok := rpcID(frame)
			if !ok || responseID != int64(id) {
				continue
			}
			if frame["error"] != nil {
				return nil, errors.New("codex model/list: provider rejected probe")
			}
			return mapField(frame, "result"), nil
		}
	}
	if _, err := request(1, "initialize", map[string]any{"clientInfo": map[string]any{"name": "riido-model-catalog", "version": "1"}}); err != nil {
		return nil, err
	}
	if err := encoder.Encode(map[string]any{"method": "initialized"}); err != nil {
		return nil, err
	}
	var models []ModelInfo
	seen := map[string]bool{}
	cursor := ""
	for page := range 5 {
		params := map[string]any{"limit": 100}
		if cursor != "" {
			params["cursor"] = cursor
		}
		result, err := request(page+2, "model/list", params)
		if err != nil {
			return nil, err
		}
		data, ok := result["data"].([]any)
		if !ok {
			return nil, errors.New("codex model/list: invalid catalog response")
		}
		for _, value := range data {
			model, _ := value.(map[string]any)
			id := stringField(model, "model")
			if id == "" {
				id = stringField(model, "id")
			}
			if id == "" || seen[id] {
				continue
			}
			seen[id] = true
			label := stringField(model, "displayName")
			if label == "" {
				label = id
			}
			isDefault, _ := model["isDefault"].(bool)
			models = append(models, ModelInfo{ModelID: id, Label: label, IsDefault: isDefault})
		}
		cursor = stringField(result, "nextCursor")
		if cursor == "" {
			return models, nil
		}
	}
	return nil, errors.New("codex model/list: catalog exceeded page bound")
}
