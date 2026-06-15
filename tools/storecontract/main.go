package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
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
