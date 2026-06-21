package main

import (
	"fmt"
	"path/filepath"
)

func run(opts options) error {
	m, err := loadManifest(opts.Manifest)
	if err != nil {
		return err
	}
	if opts.WriteDoc {
		if err := writeDoc(m.GeneratedDoc, m); err != nil {
			return err
		}
	}
	if opts.CheckDoc {
		if err := checkDoc(m.GeneratedDoc, m); err != nil {
			return err
		}
	}
	if opts.EvidenceOut == "" && (opts.WriteDoc || opts.CheckDoc) {
		return nil
	}
	report := buildReport(opts.EvidenceDir, m)
	if opts.EvidenceOut != "" {
		if err := writeJSON(opts.EvidenceOut, report); err != nil {
			return err
		}
	}
	if report.Status != statusVerified {
		return fmt.Errorf("self-improvement evidence invalid: %v", report.Problems)
	}
	return nil
}

func buildReport(dir string, m manifest) report {
	out := newReport(m)
	for _, item := range m.Required {
		path := filepath.Join(dir, item.File)
		data, err := readEvidence(path)
		if err != nil {
			out.Problems = append(out.Problems, err.Error())
			continue
		}
		checks, problems := evaluate(item, data)
		out.Checks = append(out.Checks, checks...)
		out.Problems = append(out.Problems, problems...)
	}
	for _, check := range out.Checks {
		if check.Status == statusVerified {
			out.PassingCount++
		}
	}
	out.CheckCount = len(out.Checks)
	out.ProblemCount = len(out.Problems)
	out.VerifiedCount = countVerifiedEvidence(out.Checks, m.Required)
	if out.ProblemCount > 0 {
		out.Status = statusFailed
	}
	return out
}
