package main

import (
	"encoding/json"
	"os"
	"strings"
)

type storageState struct {
	Cookies []storageCookie `json:"cookies"`
	Origins []storageOrigin `json:"origins"`
}

type storageCookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type storageOrigin struct {
	LocalStorage []storageEntry `json:"localStorage"`
}

type storageEntry struct {
	Name string `json:"name"`
}

func hydrateConfigFromStorage(cfg config) {
	state, ok := loadStorageState(*cfg.storageState)
	if !ok {
		return
	}
	if *cfg.apiToken == "" {
		*cfg.apiToken = stateToken(state)
	}
	if *cfg.workspaceID == "" {
		*cfg.workspaceID = stateWorkspaceID(state)
	}
}

func loadStorageState(path string) (storageState, bool) {
	if strings.TrimSpace(path) == "" {
		return storageState{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return storageState{}, false
	}
	var state storageState
	if json.Unmarshal(data, &state) != nil {
		return storageState{}, false
	}
	return state, true
}

func stateToken(state storageState) string {
	for _, cookie := range state.Cookies {
		if cookie.Name == "token" {
			return strings.TrimSpace(cookie.Value)
		}
	}
	return ""
}
