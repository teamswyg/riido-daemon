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
	writeStringSliceValues(&buf, values)
	buf.WriteString("}")
	return buf.String()
}

func writeStringSliceValues(buf *bytes.Buffer, values []string) {
	for i, value := range values {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(jsonLiteral(value))
	}
}
