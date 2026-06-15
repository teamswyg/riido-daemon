package main

import (
	"fmt"
	"sort"
	"strings"
)

func validateChannels(channels []channel) []string {
	required := map[string]bool{
		"developer-id":  false,
		"mac-app-store": false,
		"msix-sideload": false,
		"msix-store":    false,
	}
	var problems []string
	for _, item := range channels {
		if _, ok := required[item.ID]; ok {
			required[item.ID] = true
		}
		if strings.TrimSpace(item.ID) == "" {
			problems = append(problems, "channel id is required")
		}
		if strings.TrimSpace(item.Platform) == "" {
			problems = append(problems, fmt.Sprintf("channel %q platform is required", item.ID))
		}
		if strings.TrimSpace(item.Status) == "" {
			problems = append(problems, fmt.Sprintf("channel %q status is required", item.ID))
		}
		if strings.TrimSpace(item.RuntimeRole) == "" {
			problems = append(problems, fmt.Sprintf("channel %q runtime_role is required", item.ID))
		}
		if strings.TrimSpace(item.BackgroundRule) == "" {
			problems = append(problems, fmt.Sprintf("channel %q background_rule is required", item.ID))
		}
		if strings.TrimSpace(item.LocalIPCTransport) == "" {
			problems = append(problems, fmt.Sprintf("channel %q local_ipc_transport is required", item.ID))
		}
		if strings.TrimSpace(item.DataRoot) == "" {
			problems = append(problems, fmt.Sprintf("channel %q data_root is required", item.ID))
		}
		if strings.TrimSpace(item.UpdateMechanism) == "" {
			problems = append(problems, fmt.Sprintf("channel %q update_mechanism is required", item.ID))
		}
		if len(item.RequiredSurfaces) == 0 {
			problems = append(problems, fmt.Sprintf("channel %q required_surfaces must not be empty", item.ID))
		}
		if len(item.ForbiddenSurfaces) == 0 {
			problems = append(problems, fmt.Sprintf("channel %q forbidden_surfaces must not be empty", item.ID))
		}
		problems = append(problems, requireForbiddenSurfaces(
			item,
			"bundled-provider-cli",
			"silent-provider-install",
			"external-tcp-listener",
			"arbitrary-home-scan",
		)...)
		problems = append(problems, validateDeveloperIDSurfaces(item)...)
		problems = append(problems, validateMacAppStoreSurfaces(item)...)
		problems = append(problems, validateMSIXSurfaces(item)...)
	}
	for id, seen := range required {
		if !seen {
			problems = append(problems, fmt.Sprintf("required channel %q is missing", id))
		}
	}
	sort.Strings(problems)
	return problems
}

func validateDeveloperIDSurfaces(item channel) []string {
	if item.ID != "developer-id" {
		return nil
	}
	var problems []string
	problems = append(problems, requireRuntimeContract(
		item,
		"local-helper-broker",
		"explicit-consent",
		"unix-socket",
		"user-application-support",
		"self-managed",
	)...)
	problems = append(problems, requireRequiredSurfaces(
		item,
		"developer-id-signing",
		"notarization",
		"user-consented-background-helper",
		"local-only-ipc",
	)...)
	return problems
}

func validateMacAppStoreSurfaces(item channel) []string {
	if item.ID != "mac-app-store" {
		return nil
	}
	var problems []string
	problems = append(problems, requireRuntimeContract(
		item,
		"sandboxed-login-item-helper",
		"explicit-consent-and-store-review",
		"unix-socket",
		"app-group-or-sandbox-container",
		"app-store-managed",
	)...)
	problems = append(problems, requireRequiredSurfaces(
		item,
		"app-sandbox",
		"app-group-or-container-ipc",
		"security-scoped-workspace-grant",
		"service-management-login-item-consent",
		"helper-purpose-review-note",
		"app-sandbox-entitlement-review-notes",
		"app-store-managed-updates",
		"privacy-policy",
		"review-demo-mode",
	)...)
	problems = append(problems, requireRequiredSurfaces(item, storeReviewSubmissionRequiredSurfaces...)...)
	problems = append(problems, requireForbiddenSurfaces(
		item,
		"direct-launchagent-install",
		"self-updater",
		"third-party-installer",
		"shared-location-code-install",
		"standalone-code-download",
		"root-privilege-escalation",
	)...)
	return problems
}
