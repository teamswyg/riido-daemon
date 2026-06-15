package main

import (
	"fmt"
	"os"
	"strings"
)

func validateMSIXSurfaces(item channel) []string {
	switch item.ID {
	case "msix-sideload":
		var problems []string
		problems = append(problems, requireRuntimeContract(
			item,
			"msix-packaged-helper-broker",
			"explicit-consent",
			"windows-named-pipe",
			"windows-package-local-data",
			"self-managed",
		)...)
		problems = append(problems, requireRequiredSurfaces(
			item,
			"signed-msix-package",
			"package-identity",
			"windows-desktop-target-device-family",
			"named-pipe-local-ipc",
			"package-local-data",
			"user-consented-background-helper",
		)...)
		problems = append(problems, requireForbiddenSurfaces(
			item,
			"windows-service-default",
		)...)
		return problems
	case "msix-store":
		var problems []string
		problems = append(problems, requireRuntimeContract(
			item,
			"msix-packaged-full-trust-helper-tray",
			"explicit-consent-and-store-review",
			"windows-named-pipe",
			"windows-package-local-data",
			"store-managed",
		)...)
		problems = append(problems, requireRequiredSurfaces(
			item,
			"package-identity",
			"windows-desktop-target-device-family",
			"named-pipe-local-ipc",
			"package-local-data",
			"runfulltrust-review-note",
			"store-managed-updates",
			"partner-center-review-notes",
			"review-demo-mode",
			"privacy-policy",
		)...)
		problems = append(problems, requireRequiredSurfaces(item, storeReviewSubmissionRequiredSurfaces...)...)
		problems = append(problems, requireForbiddenSurfaces(
			item,
			"windows-service-default",
			"self-updater",
		)...)
		return problems
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

func validateRequiredDocs(repoRoot string, docs []string) []string {
	var problems []string
	for _, doc := range docs {
		path := resolvePath(repoRoot, doc)
		info, err := os.Stat(path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("required doc missing: %s", doc))
			continue
		}
		if info.IsDir() {
			problems = append(problems, fmt.Sprintf("required doc is a directory: %s", doc))
		}
	}
	return problems
}

func validateRequiredNoticeTerms(repoRoot string, terms []string) []string {
	if len(terms) == 0 {
		return nil
	}
	path := resolvePath(repoRoot, "NOTICE.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("read NOTICE.md: %v", err)}
	}
	text := string(data)
	var problems []string
	for _, term := range terms {
		trimmed := strings.TrimSpace(term)
		if trimmed == "" {
			continue
		}
		if !strings.Contains(text, trimmed) {
			problems = append(problems, fmt.Sprintf("NOTICE.md must include %q", trimmed))
		}
	}
	return problems
}
