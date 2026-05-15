package gen

import (
	"testing"

	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

func TestTsType(t *testing.T) {
	tests := []struct {
		name     string
		schema   *openapi.RawSchemaOrRef
		expected string
	}{
		{
			"string type",
			&openapi.RawSchemaOrRef{Schema: openapi.Schema{Type: "string"}},
			"string",
		},
		{
			"integer type",
			&openapi.RawSchemaOrRef{Schema: openapi.Schema{Type: "integer"}},
			"number",
		},
		{
			"number type",
			&openapi.RawSchemaOrRef{Schema: openapi.Schema{Type: "number"}},
			"number",
		},
		{
			"boolean type",
			&openapi.RawSchemaOrRef{Schema: openapi.Schema{Type: "boolean"}},
			"boolean",
		},
		{
			"array type",
			&openapi.RawSchemaOrRef{
				Schema: openapi.Schema{
					Type:  "array",
					Items: &openapi.RawSchemaOrRef{Schema: openapi.Schema{Type: "string"}},
				},
			},
			"Array<string>",
		},
		{
			"empty schema",
			&openapi.RawSchemaOrRef{},
			"any",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TsType(tt.schema, nil, nil, "")
			if got != tt.expected {
				t.Errorf("TsType(%+v) = %q, want %q", tt.schema, got, tt.expected)
			}
		})
	}
}

func TestIsArraySchema(t *testing.T) {
	tests := []struct {
		name     string
		schema   *openapi.Schema
		expected bool
	}{
		{
			"array",
			&openapi.Schema{Type: "array"},
			true,
		},
		{
			"not array",
			&openapi.Schema{Type: "string"},
			false,
		},
		{
			"no type",
			&openapi.Schema{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsArraySchema(tt.schema)
			if got != tt.expected {
				t.Errorf("IsArraySchema(%+v) = %v, want %v", tt.schema, got, tt.expected)
			}
		})
	}
}

func TestEscapeJS(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello\nworld", "hello\\nworld"},
		{"tab\there", "tab\\there"},
		{"back\\slash", "back\\\\slash"},
		{"", ""},
		{"mixed'chars\nhere\tok", "mixed\\'chars\\nhere\\tok"},
	}
	for _, tt := range tests {
		got := EscapeJS(tt.input)
		if got != tt.expected {
			t.Errorf("EscapeJS(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestTsComments(t *testing.T) {
	tests := []struct {
		name        string
		description string
		level       int
		deprecated  bool
		expected    string
	}{
		{
			"no description",
			"", 0, false,
			"",
		},
		{
			"simple description",
			"A pet", 0, false,
			"\n/**\n * A pet\n */\n",
		},
		{
			"indented",
			"prop", 1, false,
			"\n  /**\n   * prop\n   */\n  ",
		},
		{
			"deprecated",
			"Old field", 0, true,
			"\n/**\n * Old field\n *\n * @deprecated\n */\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TsComments(tt.description, tt.level, tt.deprecated)
			if got != tt.expected {
				t.Errorf("TsComments(%q, %d, %v) =\n%q\nwant:\n%q",
					tt.description, tt.level, tt.deprecated, got, tt.expected)
			}
		})
	}
}
