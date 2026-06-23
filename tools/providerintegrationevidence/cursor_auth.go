package main

import (
	"context"
	"os/exec"
	"strings"
)

type cursorAuthPreflight struct {
	InteractiveLoginProbe string `json:"interactive_login_probe"`
	InteractiveLoggedIn   bool   `json:"interactive_logged_in"`
	HeadlessAPIKeyEnv     string `json:"headless_api_key_env"`
	HeadlessAPIKeyPresent bool   `json:"headless_api_key_present"`
}

func cursorInteractiveLoggedIn(executable string) bool {
	if executable == "" {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), providerVersionTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, executable, "about").CombinedOutput()
	if err != nil {
		return false
	}
	return !strings.Contains(strings.ToLower(string(out)), "not logged in")
}
