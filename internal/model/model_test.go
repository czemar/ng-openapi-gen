package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root")
		}
		dir = parent
	}
}

func testSpec(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(findProjectRoot(t), "test", name)
}

func defaultOpts() *config.Options {
	return &config.Options{
		FetchTimeout: 20000,
		DefaultTag:   "Api",
		EnumStyle:    "alias",
	}
}

func TestNewEnumValue(t *testing.T) {
	t.Run("string enum", func(t *testing.T) {
		ev := NewEnumValue("string", "VANILLA", "Vanilla flavor", "vanilla", defaultOpts())
		if ev.Name != "VANILLA" {
			t.Errorf("Name = %q, want %q", ev.Name, "VANILLA")
		}
		if ev.Value != "'vanilla'" {
			t.Errorf("Value = %q, want %q", ev.Value, "'vanilla'")
		}
		if ev.Description != "Vanilla flavor" {
			t.Errorf("Description = %q, want %q", ev.Description, "Vanilla flavor")
		}
		if ev.Type != "string" {
			t.Errorf("Type = %q, want %q", ev.Type, "string")
		}
	})

	t.Run("number enum", func(t *testing.T) {
		ev := NewEnumValue("integer", "", "", 42, defaultOpts())
		if ev.Value != "42" {
			t.Errorf("Value = %q, want %q", ev.Value, "42")
		}
	})

	t.Run("no name falls back to enum name", func(t *testing.T) {
		ev := NewEnumValue("string", "", "", "hello-world", defaultOpts())
		if ev.Name == "" {
			t.Errorf("Name should not be empty when x-enum-names is missing")
		}
	})

	t.Run("empty name uses underscore", func(t *testing.T) {
		ev := NewEnumValue("integer", "", "", "", defaultOpts())
		if ev.Name != "_" {
			t.Errorf("Name = %q, want %q", ev.Name, "_")
		}
	})
}

func TestNewProperty(t *testing.T) {
	spec, err := openapi.ParseSpec(testSpec(t, "petstore-3.0.json"))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}

	opts := defaultOpts()
	petRaw := spec.Components.Schemas["Pet"]
	petSchema, err := openapi.ResolveSchemaRef(spec, &petRaw)
	if err != nil {
		t.Fatalf("ResolveSchemaRef: %v", err)
	}

	t.Run("string property", func(t *testing.T) {
		prop := petSchema.Properties["name"]
		p := NewProperty("Pet", "name", prop, true, opts, spec)
		if p.Name != "name" {
			t.Errorf("Name = %q, want %q", p.Name, "name")
		}
		if !p.Required {
			t.Errorf("should be required")
		}
		p.ResolveType(opts, spec, "Pet")
		if p.Type != "string" {
			t.Errorf("Type = %q, want %q", p.Type, "string")
		}
	})

	t.Run("number property", func(t *testing.T) {
		prop := petSchema.Properties["id"]
		p := NewProperty("Pet", "id", prop, true, opts, spec)
		p.ResolveType(opts, spec, "Pet")
		if p.Type != "number" {
			t.Errorf("Type = %q, want %q", p.Type, "number")
		}
	})

	t.Run("optional property", func(t *testing.T) {
		prop := petSchema.Properties["tag"]
		p := NewProperty("Pet", "tag", prop, false, opts, spec)
		if p.Required {
			t.Errorf("should not be required")
		}
		p.ResolveType(opts, spec, "Pet")
		if p.Type != "string" {
			t.Errorf("Type = %q, want %q", p.Type, "string")
		}
	})
}

func TestNewModel(t *testing.T) {
	spec, err := openapi.ParseSpec(testSpec(t, "petstore-3.0.json"))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}

	opts := defaultOpts()

	getSchema := func(name string) *openapi.Schema {
		raw := spec.Components.Schemas[name]
		s, err := openapi.ResolveSchemaRef(spec, &raw)
		if err != nil {
			t.Fatalf("ResolveSchemaRef(%s): %v", name, err)
		}
		return s
	}

	t.Run("object model", func(t *testing.T) {
		m := NewModel(spec, "Pet", getSchema("Pet"), opts)
		if m.Name != "Pet" {
			t.Errorf("Name = %q, want %q", m.Name, "Pet")
		}
		if !m.IsObject {
			t.Errorf("Pet should be an object")
		}
		if m.IsSimple {
			t.Errorf("Pet should not be simple")
		}
		if len(m.Properties) != 3 {
			t.Errorf("expected 3 properties, got %d", len(m.Properties))
		}
		if m.TypeName != "Pet" {
			t.Errorf("TypeName = %q, want %q", m.TypeName, "Pet")
		}
		if m.FileName == "" {
			t.Errorf("FileName should not be empty")
		}
	})

	t.Run("array model (simple)", func(t *testing.T) {
		m := NewModel(spec, "Pets", getSchema("Pets"), opts)
		if !m.IsSimple {
			t.Errorf("Pets should be simple (array type alias)")
		}
		if m.SimpleType == "" {
			t.Errorf("SimpleType should not be empty")
		}
	})

	t.Run("error model", func(t *testing.T) {
		m := NewModel(spec, "Error", getSchema("Error"), opts)
		if !m.IsObject {
			t.Errorf("Error should be an object")
		}
		if len(m.Properties) != 2 {
			t.Errorf("expected 2 properties, got %d", len(m.Properties))
		}
	})

	t.Run("enum model", func(t *testing.T) {
		enumSpec, err := openapi.ParseSpec(testSpec(t, "enums.json"))
		if err != nil {
			t.Fatalf("ParseSpec: %v", err)
		}
		enumRaw := enumSpec.Components.Schemas["FlavorEnum"]
		enumSchema, err := openapi.ResolveSchemaRef(enumSpec, &enumRaw)
		if err != nil {
			t.Fatalf("ResolveSchemaRef: %v", err)
		}
		opts.EnumStyle = "alias"
		opts.EnumArray = boolPtr(true)
		m := NewModel(enumSpec, "FlavorEnum", enumSchema, opts)
		if len(m.EnumValues) == 0 {
			t.Errorf("expected enum values")
		}
		if m.EnumArrayName == "" {
			t.Errorf("EnumArrayName should not be empty")
		}
		if m.EnumArrayFileName == "" {
			t.Errorf("EnumArrayFileName should not be empty")
		}
	})

	t.Run("model with prefix and suffix", func(t *testing.T) {
		prefOpts := &config.Options{
			ModelPrefix:  "Pre",
			ModelSuffix:  "Suf",
			FetchTimeout: 20000,
			DefaultTag:   "Api",
			EnumStyle:    "alias",
		}
		m := NewModel(spec, "Pet", getSchema("Pet"), prefOpts)
		if m.TypeName != "PrePetSuf" {
			t.Errorf("TypeName = %q, want %q", m.TypeName, "PrePetSuf")
		}
	})
}

func TestNewModelAllTypes(t *testing.T) {
	spec, err := openapi.ParseSpec(testSpec(t, "all-types.json"))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	opts := defaultOpts()
	opts.EnumStyle = "pascal"

	getSchema := func(name string) *openapi.Schema {
		raw := spec.Components.Schemas[name]
		s, err := openapi.ResolveSchemaRef(spec, &raw)
		if err != nil {
			t.Fatalf("ResolveSchemaRef(%s): %v", name, err)
		}
		return s
	}

	t.Run("ref enum", func(t *testing.T) {
		m := NewModel(spec, "RefEnum", getSchema("RefEnum"), opts)
		if len(m.EnumValues) == 0 {
			t.Errorf("RefEnum should have enum values")
		}
	})

	t.Run("one of", func(t *testing.T) {
		m := NewModel(spec, "Shape", getSchema("Shape"), opts)
		if !m.IsObject && !m.IsSimple {
			t.Errorf("Shape should be an object or simple")
		}
	})

	t.Run("namespaced model", func(t *testing.T) {
		m := NewModel(spec, "a.b.RefObject", getSchema("a.b.RefObject"), opts)
		if m.Namespace != "a/b" {
			t.Errorf("Namespace = %q, want %q", m.Namespace, "a/b")
		}
		if m.FileName != "a/b/ref-object" {
			t.Errorf("FileName = %q, want %q", m.FileName, "a/b/ref-object")
		}
	})
}

func boolPtr(b bool) *bool {
	return &b
}
