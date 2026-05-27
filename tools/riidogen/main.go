package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"text/template"
)

type nativeConfigPlanCatalogSpec struct {
	SchemaVersion         string                         `json:"schema_version"`
	ManifestSchemaVersion string                         `json:"manifest_schema_version"`
	Default               nativeConfigProviderPlanSpec   `json:"default"`
	Providers             []nativeConfigProviderPlanSpec `json:"providers"`
	SpecPath              string
}

type nativeConfigProviderPlanSpec struct {
	ProviderKind           string   `json:"provider_kind,omitempty"`
	PrimaryInstructionFile string   `json:"primary_instruction_file"`
	ManifestFile           string   `json:"manifest_file"`
	HookMode               string   `json:"hook_mode"`
	ConfigHomeDir          string   `json:"config_home_dir,omitempty"`
	ProviderSettingsFiles  []string `json:"provider_settings_files,omitempty"`
	HookFiles              []string `json:"hook_files,omitempty"`
	ExtraFiles             []string `json:"extra_files,omitempty"`
}

func main() {
	kind := flag.String("kind", "", "generator kind")
	specPath := flag.String("spec", "", "spec JSON path")
	templatePath := flag.String("template", "", "Go template path")
	outPath := flag.String("out", "", "output path")
	flag.Parse()

	if err := run(*kind, *specPath, *templatePath, *outPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(kind string, specPath string, templatePath string, outPath string) error {
	if kind != "native-config-plan" {
		return fmt.Errorf("riidogen: unsupported kind %q", kind)
	}
	if specPath == "" || templatePath == "" || outPath == "" {
		return errors.New("riidogen: -spec, -template, and -out are required")
	}

	spec, err := loadNativeConfigPlan(specPath)
	if err != nil {
		return err
	}
	rendered, err := renderSpec(spec, templatePath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("riidogen: create output dir: %w", err)
	}
	if err := os.WriteFile(outPath, rendered, 0o644); err != nil {
		return fmt.Errorf("riidogen: write output: %w", err)
	}
	return nil
}

func loadNativeConfigPlan(path string) (nativeConfigPlanCatalogSpec, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nativeConfigPlanCatalogSpec{}, fmt.Errorf("riidogen: read native config plan: %w", err)
	}
	var spec nativeConfigPlanCatalogSpec
	if err := json.Unmarshal(body, &spec); err != nil {
		return nativeConfigPlanCatalogSpec{}, fmt.Errorf("riidogen: decode native config plan: %w", err)
	}
	spec.SpecPath = filepath.Base(path)
	if err := validateNativeConfigPlan(spec); err != nil {
		return nativeConfigPlanCatalogSpec{}, err
	}
	sort.Slice(spec.Providers, func(i, j int) bool {
		return spec.Providers[i].ProviderKind < spec.Providers[j].ProviderKind
	})
	return spec, nil
}

func validateNativeConfigPlan(spec nativeConfigPlanCatalogSpec) error {
	if spec.SchemaVersion == "" {
		return errors.New("riidogen: native config plan schema_version is required")
	}
	if spec.ManifestSchemaVersion == "" {
		return errors.New("riidogen: native config manifest schema version is required")
	}
	if spec.Default.PrimaryInstructionFile == "" || spec.Default.ManifestFile == "" || spec.Default.HookMode == "" {
		return errors.New("riidogen: native config default plan is incomplete")
	}
	seen := map[string]struct{}{}
	for _, provider := range spec.Providers {
		if provider.ProviderKind == "" {
			return errors.New("riidogen: provider_kind is required")
		}
		if _, ok := seen[provider.ProviderKind]; ok {
			return fmt.Errorf("riidogen: duplicate provider_kind %q", provider.ProviderKind)
		}
		seen[provider.ProviderKind] = struct{}{}
		if provider.PrimaryInstructionFile == "" || provider.ManifestFile == "" || provider.HookMode == "" {
			return fmt.Errorf("riidogen: provider %q plan is incomplete", provider.ProviderKind)
		}
	}
	return nil
}

func renderSpec(data any, templatePath string) ([]byte, error) {
	body, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("riidogen: read template: %w", err)
	}
	tmpl, err := template.New(filepath.Base(templatePath)).Funcs(template.FuncMap{
		"json":          jsonLiteral,
		"goStringSlice": goStringSliceLiteral,
	}).Parse(string(body))
	if err != nil {
		return nil, fmt.Errorf("riidogen: parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("riidogen: execute template: %w", err)
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("riidogen: format generated Go: %w\n%s", err, buf.String())
	}
	return formatted, nil
}

func jsonLiteral(value any) string {
	body, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(body)
}

func goStringSliceLiteral(values []string) string {
	if len(values) == 0 {
		return "nil"
	}
	var buf bytes.Buffer
	buf.WriteString("[]string{")
	for i, value := range values {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(jsonLiteral(value))
	}
	buf.WriteString("}")
	return buf.String()
}
