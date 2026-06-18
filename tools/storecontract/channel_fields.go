package main

import (
	"fmt"
	"strings"
)

func validateChannelFields(item channel) []string {
	var problems []string
	problems = appendRequiredString(problems, item.ID, "channel id is required")
	problems = appendRequiredField(problems, item.ID, "platform", item.Platform)
	problems = appendRequiredField(problems, item.ID, "status", item.Status)
	problems = appendRequiredField(problems, item.ID, "runtime_role", item.RuntimeRole)
	problems = appendRequiredField(problems, item.ID, "background_rule", item.BackgroundRule)
	problems = appendRequiredField(problems, item.ID, "local_ipc_transport", item.LocalIPCTransport)
	problems = appendRequiredField(problems, item.ID, "data_root", item.DataRoot)
	problems = appendRequiredField(problems, item.ID, "update_mechanism", item.UpdateMechanism)
	problems = appendRequiredList(problems, item.ID, "required_surfaces", item.RequiredSurfaces)
	return appendRequiredList(problems, item.ID, "forbidden_surfaces", item.ForbiddenSurfaces)
}

func appendRequiredField(problems []string, id, field, value string) []string {
	message := fmt.Sprintf("channel %q %s is required", id, field)
	return appendRequiredString(problems, value, message)
}

func appendRequiredString(problems []string, value, message string) []string {
	if strings.TrimSpace(value) == "" {
		return append(problems, message)
	}
	return problems
}

func appendRequiredList(problems []string, id, field string, items []string) []string {
	if len(items) == 0 {
		return append(problems, fmt.Sprintf("channel %q %s must not be empty", id, field))
	}
	return problems
}
