package main

import (
	"context"
	"fmt"
	"os"
)

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

func run(_ context.Context, opts options) error {
	manifest, err := loadManifest(opts.Repo, opts.Manifest)
	if err != nil {
		return err
	}
	problems, mappings, coverage := validate(opts.Repo, manifest)
	rendered := render(manifest)
	if err := maybeWriteDoc(opts, manifest, rendered); err != nil {
		return err
	}
	problems = append(problems, checkDoc(opts, manifest, rendered)...)
	if opts.EvidenceOut != "" {
		if err := writeJSON(opts.EvidenceOut, buildEvidence(manifest, problems, mappings, coverage)); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("provider-draft-mapping: clean")
	return nil
}

func maybeWriteDoc(opts options, manifest Manifest, rendered string) error {
	if !opts.WriteDoc {
		return nil
	}
	path, err := cleanRepoPath(opts.Repo, manifest.GeneratedDoc)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(rendered), 0o644)
}

func checkDoc(opts options, manifest Manifest, rendered string) []problem {
	if !opts.CheckDoc {
		return nil
	}
	path, err := cleanRepoPath(opts.Repo, manifest.GeneratedDoc)
	if err != nil {
		return []problem{{err.Error()}}
	}
	actual, err := os.ReadFile(path)
	if err != nil {
		return []problem{{err.Error()}}
	}
	if string(actual) != rendered {
		return []problem{{"generated doc drift: run go run ./tools/providerdraftmapping -write-doc"}}
	}
	return nil
}
