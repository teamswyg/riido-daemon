package userapi

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"time"
)

// Typed JSON decoding matches keys case-insensitively and accepts null scalar
// values. Validate exact DTO field names first, without case-folding dynamic
// snapshot IDs. The preceding syntax walk bounds nesting and rejects duplicates.
func threadJSONShape(raw []byte, kind reflect.Type) bool {
	if kind == reflect.TypeFor[json.RawMessage]() {
		return true // execution_context is authorized server metadata, kept verbatim.
	}
	isNull := bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
	if kind.Kind() == reflect.Pointer {
		return isNull || threadJSONShape(raw, kind.Elem())
	}
	if kind == reflect.TypeFor[time.Time]() {
		return !isNull
	}
	switch kind.Kind() {
	case reflect.Struct:
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil || fields == nil {
			return false
		}
		allowed := map[string]reflect.Type{}
		threadJSONFields(kind, allowed)
		for name, value := range fields {
			field, ok := allowed[name]
			if !ok || !threadJSONShape(value, field) {
				return false
			}
		}
	case reflect.Slice:
		var values []json.RawMessage
		if json.Unmarshal(raw, &values) != nil {
			return false
		}
		for _, value := range values {
			if !threadJSONShape(value, kind.Elem()) {
				return false
			}
		}
	case reflect.Map:
		var values map[string]json.RawMessage
		if json.Unmarshal(raw, &values) != nil {
			return false
		}
		for _, value := range values {
			if !threadJSONShape(value, kind.Elem()) {
				return false
			}
		}
	default:
		return !isNull
	}
	return true
}

func threadJSONFields(kind reflect.Type, fields map[string]reflect.Type) {
	for field := range kind.Fields() {
		if field.Anonymous {
			threadJSONFields(field.Type, fields)
			continue
		}
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name != "" && name != "-" {
			fields[name] = field.Type
		}
	}
}
