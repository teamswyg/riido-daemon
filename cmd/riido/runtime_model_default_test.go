package main

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"testing"
)

func TestRuntimeModelUnknownDefaultDoesNotPickFirst(t *testing.T) {
	for _, id := range []string{"", "missing"} {
		for _, model := range normalizeRuntimeModels([]runtimeactor.RuntimeModel{{ModelID: "a"}, {ModelID: "b"}}, id) {
			if model.IsDefault {
				t.Fatalf("invented default %+v", model)
			}
		}
	}
	models := parseCursorRuntimeModelList([]byte("a - Alpha\nauto - Auto (current, default)\n"), "")
	if len(models) != 2 || models[0].IsDefault || !models[1].IsDefault || models[1].Label != "Auto" {
		t.Fatalf("models=%+v", models)
	}
}
