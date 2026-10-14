package main

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	cursorprovider "github.com/teamswyg/riido-daemon/internal/provider/cursor"
)

const cursorProviderName = cursorprovider.Name

func cursorRuntimeModelsFromCommand(defaultID string) []runtimeactor.RuntimeModel {
	body := runtimeModelCommandOutput(cursorprovider.DefaultExecutable, "models")
	return parseCursorRuntimeModelList(body, defaultID)
}
