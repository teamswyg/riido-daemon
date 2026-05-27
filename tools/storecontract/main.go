package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	contractSchemaVersion = "riido-store-distribution-contract.v1"
	checkSchemaVersion    = "riido-store-distribution-contract-check.v1"
)

var storeReviewSubmissionRequiredSurfaces = []string{
	"demo-review-account",
	"privacy-metadata-allowlist",
	"provider-non-bundling-review-note",
}

type contract struct {
	SchemaVersion            string    `json:"schema_version"`
	Product                  string    `json:"product"`
	ProviderCLIBundling      string    `json:"provider_cli_bundling"`
	ExternalProviderCLINames []string  `json:"external_provider_cli_names"`
	StoreArtifactRoots       []string  `json:"store_artifact_roots"`
	RequiredDocs             []string  `json:"required_docs"`
	RequiredNoticeTerms      []string  `json:"required_notice_terms"`
	Channels                 []channel `json:"channels"`
}

type channel struct {
	ID                string   `json:"id"`
	Platform          string   `json:"platform"`
	Status            string   `json:"status"`
	RuntimeRole       string   `json:"runtime_role"`
	BackgroundRule    string   `json:"background_rule"`
	LocalIPCTransport string   `json:"local_ipc_transport"`
	DataRoot          string   `json:"data_root"`
	UpdateMechanism   string   `json:"update_mechanism"`
	RequiredSurfaces  []string `json:"required_surfaces"`
	ForbiddenSurfaces []string `json:"forbidden_surfaces"`
}

type checkResult struct {
	SchemaVersion      string   `json:"schema_version"`
	ContractPath       string   `json:"contract_path"`
	Product            string   `json:"product"`
	Status             string   `json:"status"`
	Channels           []string `json:"channels"`
	StoreArtifactRoots []string `json:"store_artifact_roots"`
	Errors             []string `json:"errors,omitempty"`
}

func main() {
	contractPath := flag.String("contract", "packaging/store/riido_daemon_store_distribution.riido.json", "store distribution contract path")
	repoRoot := flag.String("repo", ".", "repository root")
	outPath := flag.String("out", "", "optional JSON check output path")
	flag.Parse()

	result, err := run(*repoRoot, *contractPath)
	if *outPath != "" {
		if writeErr := writeJSON(*outPath, result); writeErr != nil {
			fmt.Fprintln(os.Stderr, writeErr)
			os.Exit(1)
		}
	}
	if err != nil {
		for _, message := range result.Errors {
			fmt.Fprintln(os.Stderr, message)
		}
		os.Exit(1)
	}
	fmt.Println("store-distribution-contract: clean")
}

func run(repoRoot, contractPath string) (checkResult, error) {
	result := checkResult{
		SchemaVersion: checkSchemaVersion,
		ContractPath:  contractPath,
		Status:        "failed",
	}

	root, err := filepath.Abs(repoRoot)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("resolve repo root: %v", err))
		return result, errors.New("invalid repo root")
	}

	contractFile := resolvePath(root, contractPath)
	loaded, err := loadContract(contractFile)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}
	result.Product = loaded.Product
	result.StoreArtifactRoots = append(result.StoreArtifactRoots, loaded.StoreArtifactRoots...)
	for _, item := range loaded.Channels {
		result.Channels = append(result.Channels, item.ID)
	}
	sort.Strings(result.Channels)

	var problems []string
	problems = append(problems, validateContractShape(loaded)...)
	problems = append(problems, validateRequiredDocs(root, loaded.RequiredDocs)...)
	problems = append(problems, validateRequiredNoticeTerms(root, loaded.RequiredNoticeTerms)...)
	problems = append(problems, scanArtifactRoots(root, loaded.StoreArtifactRoots, loaded.ExternalProviderCLINames)...)
	if len(problems) > 0 {
		result.Errors = problems
		return result, errors.New("store distribution contract failed")
	}

	result.Status = "passed"
	return result, nil
}

func loadContract(path string) (contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return contract{}, fmt.Errorf("read contract %s: %w", path, err)
	}
	var loaded contract
	if err := json.Unmarshal(data, &loaded); err != nil {
		return contract{}, fmt.Errorf("decode contract %s: %w", path, err)
	}
	return loaded, nil
}

func validateContractShape(loaded contract) []string {
	var problems []string
	if loaded.SchemaVersion != contractSchemaVersion {
		problems = append(problems, fmt.Sprintf("schema_version must be %q", contractSchemaVersion))
	}
	if strings.TrimSpace(loaded.Product) == "" {
		problems = append(problems, "product is required")
	}
	if loaded.ProviderCLIBundling != "forbidden" {
		problems = append(problems, `provider_cli_bundling must be "forbidden"`)
	}
	if len(loaded.ExternalProviderCLINames) == 0 {
		problems = append(problems, "external_provider_cli_names must not be empty")
	}
	for _, name := range loaded.ExternalProviderCLINames {
		if strings.TrimSpace(name) == "" || strings.ContainsAny(name, `/\`) {
			problems = append(problems, fmt.Sprintf("invalid provider CLI name %q", name))
		}
	}
	if len(loaded.StoreArtifactRoots) == 0 {
		problems = append(problems, "store_artifact_roots must not be empty")
	}
	if len(loaded.RequiredNoticeTerms) == 0 {
		problems = append(problems, "required_notice_terms must not be empty")
	}
	for _, term := range loaded.RequiredNoticeTerms {
		if strings.TrimSpace(term) == "" {
			problems = append(problems, "required_notice_terms must not include empty terms")
		}
	}
	problems = append(problems, validateChannels(loaded.Channels)...)
	return problems
}

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
		problems = append(problems, requireForbiddenSurfaces(item,
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
	problems = append(problems, requireRuntimeContract(item,
		"local-helper-broker",
		"explicit-consent",
		"unix-socket",
		"user-application-support",
		"self-managed",
	)...)
	problems = append(problems, requireRequiredSurfaces(item,
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
	problems = append(problems, requireRuntimeContract(item,
		"sandboxed-login-item-helper",
		"explicit-consent-and-store-review",
		"unix-socket",
		"app-group-or-sandbox-container",
		"app-store-managed",
	)...)
	problems = append(problems, requireRequiredSurfaces(item,
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
	problems = append(problems, requireForbiddenSurfaces(item,
		"direct-launchagent-install",
		"self-updater",
		"third-party-installer",
		"shared-location-code-install",
		"standalone-code-download",
		"root-privilege-escalation",
	)...)
	return problems
}

func validateMSIXSurfaces(item channel) []string {
	switch item.ID {
	case "msix-sideload":
		var problems []string
		problems = append(problems, requireRuntimeContract(item,
			"msix-packaged-helper-broker",
			"explicit-consent",
			"windows-named-pipe",
			"windows-package-local-data",
			"self-managed",
		)...)
		problems = append(problems, requireRequiredSurfaces(item,
			"signed-msix-package",
			"package-identity",
			"windows-desktop-target-device-family",
			"named-pipe-local-ipc",
			"package-local-data",
			"user-consented-background-helper",
		)...)
		problems = append(problems, requireForbiddenSurfaces(item,
			"windows-service-default",
		)...)
		return problems
	case "msix-store":
		var problems []string
		problems = append(problems, requireRuntimeContract(item,
			"msix-packaged-full-trust-helper-tray",
			"explicit-consent-and-store-review",
			"windows-named-pipe",
			"windows-package-local-data",
			"store-managed",
		)...)
		problems = append(problems, requireRequiredSurfaces(item,
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
		problems = append(problems, requireForbiddenSurfaces(item,
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

func scanArtifactRoots(repoRoot string, roots []string, providerNames []string) []string {
	var problems []string
	for _, root := range roots {
		path := resolvePath(repoRoot, root)
		info, err := os.Stat(path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("store artifact root missing: %s", root))
			continue
		}
		if !info.IsDir() {
			problems = append(problems, fmt.Sprintf("store artifact root is not a directory: %s", root))
			continue
		}
		walkErr := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				problems = append(problems, fmt.Sprintf("scan %s: %v", path, err))
				return nil
			}
			if entry.IsDir() {
				return nil
			}
			if matchesProviderBinary(entry.Name(), providerNames) {
				problems = append(problems, fmt.Sprintf("provider CLI appears bundled in store artifact root: %s", path))
			}
			if hasHardcodedUserPath(path) {
				problems = append(problems, fmt.Sprintf("store artifact contains hardcoded user path: %s", path))
			}
			return nil
		})
		if walkErr != nil {
			problems = append(problems, fmt.Sprintf("scan root %s: %v", root, walkErr))
		}
	}
	return problems
}

func matchesProviderBinary(filename string, providerNames []string) bool {
	base := strings.ToLower(filename)
	ext := strings.ToLower(filepath.Ext(base))
	stem := strings.TrimSuffix(base, ext)
	executableExt := ext == "" || ext == ".exe" || ext == ".cmd" || ext == ".bat" || ext == ".ps1" || ext == ".sh"
	if !executableExt {
		return false
	}
	for _, provider := range providerNames {
		name := strings.ToLower(provider)
		if base == name || stem == name {
			return true
		}
	}
	return false
}

func hasHardcodedUserPath(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	text := string(data)
	for _, marker := range []string{"/Users/", `C:\Users\`, "C:/Users/", "~/Library/LaunchAgents", "~/Library/Application Support"} {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func contains(items []string, wanted string) bool {
	for _, item := range items {
		if item == wanted {
			return true
		}
	}
	return false
}

func resolvePath(repoRoot, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(repoRoot, path)
}

func writeJSON(path string, value checkResult) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode check output: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write check output: %w", err)
	}
	return nil
}
