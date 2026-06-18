package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

func renderSpec(data any, templatePath string) ([]byte, error) {
	body, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("riidogen: read template: %w", err)
	}
	tmpl, err := parseTemplate(templatePath, body)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("riidogen: execute template: %w", err)
	}
	return formatGeneratedSource(buf)
}

func parseTemplate(path string, body []byte) (*template.Template, error) {
	tmpl, err := template.New(filepath.Base(path)).Funcs(templateFuncMap()).Parse(string(body))
	if err != nil {
		return nil, fmt.Errorf("riidogen: parse template: %w", err)
	}
	return tmpl, nil
}

func formatGeneratedSource(buf bytes.Buffer) ([]byte, error) {
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("riidogen: format generated Go: %w\n%s", err, buf.String())
	}
	return formatted, nil
}
