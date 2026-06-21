package main

import "fmt"

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
