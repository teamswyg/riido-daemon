package main

import (
	"bytes"
	"encoding/json"
	"text/template"
)

func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"json":          jsonLiteral,
		"goStringSlice": goStringSliceLiteral,
	}
}

func jsonLiteral(value any) (string, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func goStringSliceLiteral(values []string) (string, error) {
	if len(values) == 0 {
		return "nil", nil
	}
	var buf bytes.Buffer
	buf.WriteString("[]string{")
	if err := writeStringSliceValues(&buf, values); err != nil {
		return "", err
	}
	buf.WriteString("}")
	return buf.String(), nil
}

func writeStringSliceValues(buf *bytes.Buffer, values []string) error {
	for i, value := range values {
		if i > 0 {
			buf.WriteString(", ")
		}
		literal, err := jsonLiteral(value)
		if err != nil {
			return err
		}
		buf.WriteString(literal)
	}
	return nil
}
