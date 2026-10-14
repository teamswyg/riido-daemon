package jsontest

import (
	"reflect"
	"strings"
)

// StructJSONPaths returns exported JSON field paths for a struct shape.
func StructJSONPaths(t reflect.Type) []string {
	return structJSONPaths(t, "")
}

func structJSONPaths(t reflect.Type, prefix string) []string {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	var paths []string
	for field := range t.Fields() {
		if field.PkgPath != "" {
			continue
		}
		name := jsonTagName(field)
		if name == "" {
			continue
		}
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		paths = append(paths, path)

		fieldType := field.Type
		if fieldType.Kind() == reflect.Slice {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct && fieldType.PkgPath() == t.PkgPath() {
			paths = append(paths, structJSONPaths(fieldType, path)...)
		}
	}
	return paths
}

func jsonTagName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return ""
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		name = field.Name
	}
	return name
}
