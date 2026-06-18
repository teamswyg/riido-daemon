package main

import "fmt"

func validateMSIXSurfaces(item channel) []string {
	switch item.ID {
	case "msix-sideload":
		return validateMSIXSideloadSurfaces(item)
	case "msix-store":
		return validateMSIXStoreSurfaces(item)
	default:
		return nil
	}
}

func requireRuntimeContract(item channel, role, backgroundRule, ipc, dataRoot, updateMechanism string) []string {
	var problems []string
	if item.RuntimeRole != role {
		problems = append(problems, fmt.Sprintf("channel %q runtime_role must be %s", item.ID, role))
	}
	if item.BackgroundRule != backgroundRule {
		problems = append(problems, fmt.Sprintf("channel %q background_rule must be %s", item.ID, backgroundRule))
	}
	if item.LocalIPCTransport != ipc {
		problems = append(problems, fmt.Sprintf("channel %q local_ipc_transport must be %s", item.ID, ipc))
	}
	if item.DataRoot != dataRoot {
		problems = append(problems, fmt.Sprintf("channel %q data_root must be %s", item.ID, dataRoot))
	}
	if item.UpdateMechanism != updateMechanism {
		problems = append(problems, fmt.Sprintf("channel %q update_mechanism must be %s", item.ID, updateMechanism))
	}
	return problems
}

func requireRequiredSurfaces(item channel, required ...string) []string {
	var problems []string
	for _, surface := range required {
		if !contains(item.RequiredSurfaces, surface) {
			problems = append(problems, fmt.Sprintf("channel %q must require %s", item.ID, surface))
		}
	}
	return problems
}

func requireForbiddenSurfaces(item channel, forbidden ...string) []string {
	var problems []string
	for _, surface := range forbidden {
		if !contains(item.ForbiddenSurfaces, surface) {
			problems = append(problems, fmt.Sprintf("channel %q must forbid %s", item.ID, surface))
		}
	}
	return problems
}
