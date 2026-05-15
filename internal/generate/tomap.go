package generate

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// toMap converts any value to map[string]any using reflection.
// It handles circular references by tracking visited pointers.
func toMap(v any) map[string]any {
	visited := make(map[uintptr]bool)
	result, _ := convertValue(reflect.ValueOf(v), visited, 0)
	if m, ok := result.(map[string]any); ok {
		return m
	}
	return make(map[string]any)
}

func convertValue(v reflect.Value, visited map[uintptr]bool, depth int) (any, bool) {
	if depth > 20 {
		return nil, false
	}

	if !v.IsValid() {
		return nil, false
	}

	// Unwrap interface
	for v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil, false
		}
		v = v.Elem()
		if !v.IsValid() {
			return nil, false
		}
	}

	// Dereference pointers with cycle detection
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, false
		}
		ptr := v.Pointer()
		if visited[ptr] {
			return nil, false
		}
		visited[ptr] = true
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return convertStruct(v, visited, depth+1), true
	case reflect.Map:
		return convertMap(v, visited, depth+1), true
	case reflect.Slice, reflect.Array:
		return convertSlice(v, visited, depth+1), true
	case reflect.String:
		return v.String(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint(), true
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	case reflect.Bool:
		return v.Bool(), true
	default:
		if v.CanInterface() {
			return fmt.Sprintf("%v", v.Interface()), true
		}
		return nil, false
	}
}

func convertStruct(v reflect.Value, visited map[uintptr]bool, depth int) map[string]any {
	result := make(map[string]any)
	t := v.Type()

	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		name := fieldJSONName(field)
		if name == "-" {
			continue
		}

		fieldVal := v.Field(i)
		if !fieldVal.IsValid() {
			continue
		}

		converted, ok := convertValue(fieldVal, visited, depth)
		if ok {
			result[name] = converted
		}
	}

	return result
}

func convertMap(v reflect.Value, visited map[uintptr]bool, depth int) map[string]any {
	result := make(map[string]any)
	for _, key := range v.MapKeys() {
		kStr := fmt.Sprintf("%v", key.Interface())
		val, ok := convertValue(v.MapIndex(key), visited, depth)
		if ok {
			result[kStr] = val
		}
	}
	return result
}

func convertSlice(v reflect.Value, visited map[uintptr]bool, depth int) []any {
	n := v.Len()
	result := make([]any, 0, n)
	for i := range n {
		val, ok := convertValue(v.Index(i), visited, depth)
		if ok {
			result = append(result, val)
		}
	}
	return result
}

func fieldJSONName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag != "" {
		parts := strings.Split(tag, ",")
		if parts[0] != "" {
			return parts[0]
		}
		if parts[0] == "-" {
			return "-"
		}
	}
	// Default: use lowerCamelCase of Go field name
	return toLowerCamel(field.Name)
}

func toLowerCamel(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
