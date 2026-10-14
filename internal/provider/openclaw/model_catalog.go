package openclaw

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
	"github.com/teamswyg/riido-daemon/internal/process/execwindow"
)

type CatalogModel struct {
	ModelID   string
	Label     string
	IsDefault bool
}

// ModelCatalog lists native routes without running inference. CLI pins remain
// authoritative, stdout is separate from diagnostic noise, and work is bounded.
func ModelCatalog(ctx context.Context, env agentbridge.DetectEnv) ([]CatalogModel, error) {
	exe, ok := detectutil.ResolveExecutable(DefaultExecutable, envValue(env, EnvOverride))
	if !ok {
		return nil, errors.New("OpenClaw catalog executable unavailable")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "models", "list", "--json")
	cmd.WaitDelay = 100 * time.Millisecond
	execwindow.Hide(cmd)
	output := &catalogOutput{}
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		return nil, errors.New("OpenClaw native model catalog unavailable")
	}
	if output.overflow {
		return nil, errors.New("OpenClaw native model catalog exceeded 64 KiB")
	}
	return parseModelCatalog(output.body)
}

func parseModelCatalog(body []byte) ([]CatalogModel, error) {
	var response struct {
		Models []struct {
			Key       string   `json:"key"`
			Name      string   `json:"name"`
			Available *bool    `json:"available"`
			Tags      []string `json:"tags"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.New("OpenClaw native model catalog invalid")
	}
	var models []CatalogModel
	seen := map[string]bool{}
	for _, row := range response.Models {
		id := strings.TrimSpace(row.Key)
		if id == "" || len(id) > 256 || seen[id] || (row.Available != nil && !*row.Available) {
			continue
		}
		seen[id] = true
		label := strings.TrimSpace(row.Name)
		if label == "" {
			label = id
		}
		if len(label) > 256 {
			label = id
		}
		model := CatalogModel{ModelID: id, Label: label}
		for _, tag := range row.Tags {
			if tag == "default" {
				model.IsDefault = true
			}
		}
		models = append(models, model)
	}
	return models, nil
}

type catalogOutput struct {
	body     []byte
	overflow bool
}

func (b *catalogOutput) Write(p []byte) (int, error) {
	n := len(p)
	remaining := 64*1024 - len(b.body)
	if len(p) > remaining {
		b.overflow = true
		p = p[:remaining]
	}
	b.body = append(b.body, p...)
	return n, nil
}
