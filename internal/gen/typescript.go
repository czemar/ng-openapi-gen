package gen

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

// TsType returns the TypeScript type for a given schema
func TsType(sRaw *openapi.RawSchemaOrRef, opts *config.Options, spec *openapi.Spec, container string) string {
	if sRaw == nil {
		return "any"
	}
	if sRaw.Ref != "" {
		schema, err := openapi.ResolveSchemaRef(spec, sRaw)
		if err != nil || schema == nil {
			return "any"
		}
		name := SimpleName(sRaw.Ref)
		if container != "" && container == name {
			return maybeAppendNull(UnqualifiedName(name, opts), isNullable(schema))
		}
		return maybeAppendNull(QualifiedName(name, opts), isNullable(schema))
	}
	return rawTsType(&sRaw.Schema, opts, spec, container)
}

// rawTsType resolves the TypeScript type from a Schema (non-ref)
func rawTsType(schema *openapi.Schema, opts *config.Options, spec *openapi.Spec, container string) string {
	// Union types (oneOf / anyOf)
	if len(schema.OneOf) > 0 || len(schema.AnyOf) > 0 {
		union := schema.OneOf
		if len(union) == 0 {
			union = schema.AnyOf
		}
		types := make([]string, len(union))
		for i, u := range union {
			u := u
			types[i] = TsType(&u, opts, spec, container)
		}
		if len(types) > 1 {
			return "(" + strings.Join(types, " | ") + ")"
		}
		return types[0]
	}

	schemaType := resolveSchemaType(schema)

	// Handle OpenAPI 3.1 union types (type array)
	if types, ok := schemaType.([]string); ok {
		return handleUnionTypes(types, schema, opts, spec, container)
	}

	typeStr, _ := schemaType.(string)

	// Array with prefix items (tuples, OpenAPI 3.1)
	if typeStr == "array" && len(schema.PrefixItems) > 0 {
		return handleTupleTypes(schema, opts, spec, container)
	}

	// Array
	if typeStr == "array" && schema.Items != nil {
		itemsType := TsType(schema.Items, opts, spec, container)
		return "Array<" + itemsType + ">"
	}

	// Intersection (allOf)
	var intersectionTypes []string
	if len(schema.AllOf) > 0 {
		for _, a := range schema.AllOf {
			a := a
			intersectionTypes = append(intersectionTypes, TsType(&a, opts, spec, container))
		}
	}

	// Object
	if typeStr == "object" || len(schema.Properties) > 0 {
		result := buildObjectType(schema, opts, spec, container)
		intersectionTypes = append(intersectionTypes, result)
	}

	if len(intersectionTypes) > 0 {
		return strings.Join(intersectionTypes, " & ")
	}

	// Inline enum
	if len(schema.Enum) > 0 || schema.Const != nil {
		return handleInlineEnum(schema)
	}

	// Binary blob
	if typeStr == "string" && schema.Format == "binary" {
		return "Blob"
	}

	// Simple type
	if typeStr != "" {
		if typeStr == "integer" {
			return "number"
		}
		return typeStr
	}

	return "any"
}

func resolveSchemaType(schema *openapi.Schema) any {
	if schema.Type != nil {
		return schema.Type
	}
	return nil
}

func isNullable(schema *openapi.Schema) bool {
	if schema.Nullable {
		return true
	}
	if types, ok := schema.Type.([]any); ok {
		for _, t := range types {
			if s, ok := t.(string); ok && s == "null" {
				return true
			}
		}
	}
	return false
}

func IsArraySchema(schema *openapi.Schema) bool {
	if schema.Type != nil {
		if s, ok := schema.Type.(string); ok {
			return s == "array"
		}
		if types, ok := schema.Type.([]any); ok {
			for _, t := range types {
				if s, ok := t.(string); ok && s == "array" {
					return true
				}
			}
		}
	}
	return schema.Items != nil
}

func handleUnionTypes(types []string, schema *openapi.Schema, opts *config.Options, spec *openapi.Spec, container string) string {
	var nonNullTypes []string
	hasNull := false
	for _, t := range types {
		if t == "null" {
			hasNull = true
		} else {
			nonNullTypes = append(nonNullTypes, t)
		}
	}

	if len(nonNullTypes) > 1 {
		unionTypes := make([]string, len(nonNullTypes))
		for i, t := range nonNullTypes {
			clone := *schema
			clone.Type = t
			unionTypes[i] = rawTsType(&clone, opts, spec, container)
		}
		unique := uniqueStrings(unionTypes)
		result := strings.Join(unique, " | ")
		if len(unique) > 1 {
			result = "(" + result + ")"
		}
		if hasNull {
			result = "(" + result + " | null)"
		}
		return result
	} else if len(nonNullTypes) == 1 {
		clone := *schema
		clone.Type = nonNullTypes[0]
		result := rawTsType(&clone, opts, spec, container)
		if hasNull {
			return "(" + result + " | null)"
		}
		return result
	} else if hasNull {
		return "null"
	}
	return "any"
}

func handleTupleTypes(schema *openapi.Schema, opts *config.Options, spec *openapi.Spec, container string) string {
	tupleTypes := make([]string, len(schema.PrefixItems))
	for i, item := range schema.PrefixItems {
		item := item
		tupleTypes[i] = TsType(&item, opts, spec, container)
	}

	if schema.Items == nil {
		return "[" + strings.Join(tupleTypes, ", ") + "]"
	}

	// Check if items is false (meaning no additional items)
	additionalType := TsType(schema.Items, opts, spec, container)
	return fmt.Sprintf("[%s, ...%s[]]", strings.Join(tupleTypes, ", "), additionalType)
}

func buildObjectType(schema *openapi.Schema, opts *config.Options, spec *openapi.Spec, container string) string {
	var sb strings.Builder
	sb.WriteString("{\n")
	required := makeSet(schema.Required)

	// Sort property names for deterministic output
	propNames := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		propNames = append(propNames, name)
	}
	sort.Strings(propNames)

	for _, propName := range propNames {
		prop := schema.Properties[propName]
		// Resolve ref for description/deprecated
		propSchema, _ := openapi.ResolveSchemaRef(spec, &prop)
		if propSchema != nil && propSchema.Description != "" {
			sb.WriteString(TsComments(propSchema.Description, 0, propSchema.Deprecated))
		}
		sb.WriteString("'" + propName + "'")
		if !required[propName] {
			sb.WriteString("?")
		}
		propType := TsType(&prop, opts, spec, container)
		sb.WriteString(": " + propType + ";\n")
	}

	// Additional properties
	if schema.AdditionalProperties != nil {
		addPropsType := "any"
		if addSchema, ok := schema.AdditionalProperties.(map[string]any); ok {
			// Convert to RawSchemaOrRef
			// This is handled through templating, simplified here
			_ = addSchema
		}
		_ = addPropsType
		sb.WriteString("  [key: string]: any;\n")
	}

	sb.WriteString("}")
	return sb.String()
}

func handleInlineEnum(schema *openapi.Schema) string {
	enumValues := schema.Enum
	if len(enumValues) == 0 && schema.Const != nil {
		enumValues = []any{schema.Const}
	}

	schemaType := resolveSchemaType(schema)
	typeStr, _ := schemaType.(string)

	parts := make([]string, len(enumValues))
	for i, v := range enumValues {
		if typeStr == "number" || typeStr == "integer" || typeStr == "boolean" {
			parts[i] = fmt.Sprintf("%v", v)
		} else {
			parts[i] = "'" + EscapeJS(fmt.Sprintf("%v", v)) + "'"
		}
	}
	return strings.Join(parts, " | ")
}

func maybeAppendNull(typeStr string, nullable bool) string {
	if strings.Contains(typeStr, "null") || !nullable {
		return typeStr
	}
	if strings.Contains(typeStr, " ") {
		return "(" + typeStr + " | null)"
	}
	return typeStr + " | null"
}

func EscapeJS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

var reNonWord = regexp.MustCompile(`[^\w$]+`)

func uniqueStrings(s []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

func makeSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, item := range items {
		m[item] = true
	}
	return m
}
