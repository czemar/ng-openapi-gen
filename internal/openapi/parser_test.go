package openapi

import (
	"encoding/json"
	"testing"

	"github.com/czemar/ng-openapi-gen/internal/testutil"
)

func TestParseSpec(t *testing.T) {
	petstorePath := testutil.TestSpecPath(t, "petstore-3.0.json")
	spec, err := ParseSpec(petstorePath)
	if err != nil {
		t.Fatalf("ParseSpec(%q) failed: %v", petstorePath, err)
	}

	if spec.OpenAPI != "3.0.0" {
		t.Errorf("spec.OpenAPI = %q, want %q", spec.OpenAPI, "3.0.0")
	}
	if spec.Info == nil {
		t.Fatal("spec.Info is nil")
	}
	if spec.Info["title"] != "Swagger Petstore" {
		t.Errorf("spec.Info['title'] = %q, want %q", spec.Info["title"], "Swagger Petstore")
	}
	if spec.Info["version"] != "1.0.0" {
		t.Errorf("spec.Info['version'] = %q, want %q", spec.Info["version"], "1.0.0")
	}

	if spec.Paths == nil {
		t.Fatal("spec.Paths is nil")
	}

	expectedPaths := []string{"/pets", "/pets/{petId}"}
	for _, p := range expectedPaths {
		if spec.Paths[p] == nil {
			t.Errorf("expected path %q not found", p)
		}
	}

	petsPath := spec.Paths["/pets"]
	if petsPath.Get == nil {
		t.Error("expected GET /pets")
	}
	if petsPath.Post == nil {
		t.Error("expected POST /pets")
	}

	if spec.Components == nil {
		t.Fatal("spec.Components is nil")
	}
	if spec.Components.Schemas == nil {
		t.Fatal("spec.Components.Schemas is nil")
	}

	expectedSchemas := []string{"Pet", "Pets", "Error"}
	for _, s := range expectedSchemas {
		if _, ok := spec.Components.Schemas[s]; !ok {
			t.Errorf("expected schema %q not found", s)
		}
	}

	petSchema := spec.Components.Schemas["Pet"]
	if petSchema.Ref != "" {
		t.Errorf("Pet schema should not be a $ref, got %q", petSchema.Ref)
	}
	// Pet has no explicit "type" field — it uses properties to imply object type
	if petSchema.Properties == nil {
		t.Fatal("Pet schema properties is nil")
	}
	expectedProps := []string{"id", "name", "tag"}
	for _, prop := range expectedProps {
		if _, ok := petSchema.Properties[prop]; !ok {
			t.Errorf("expected property %q on Pet", prop)
		}
	}
	// Check id property type
	idProp := petSchema.Properties["id"]
	if idProp.Type != "integer" {
		t.Errorf("Pet.id type = %v, want %q", idProp.Type, "integer")
	}
	// Check the required array
	if len(petSchema.Required) != 2 {
		t.Errorf("Pet.required length = %d, want 2", len(petSchema.Required))
	}
}

func TestParseAllTypes(t *testing.T) {
	specPath := testutil.TestSpecPath(t, "all-types.json")

	spec, err := ParseSpec(specPath)
	if err != nil {
		t.Fatalf("ParseSpec(%q) failed: %v", specPath, err)
	}

	if spec.OpenAPI != "3.0" {
		t.Errorf("spec.OpenAPI = %q, want %q", spec.OpenAPI, "3.0")
	}
	if spec.Components == nil || spec.Components.Schemas == nil {
		t.Fatal("no schemas found")
	}

	schemaNames := []string{
		"RefEnum", "RefIntEnum", "RefNamedIntEnum",
		"Shape", "Circle", "union",
		"AdditionalProperties", "NullableObject",
		"InlineObject", "EscapedProperties",
	}
	for _, name := range schemaNames {
		if _, ok := spec.Components.Schemas[name]; !ok {
			t.Errorf("expected schema %q not found", name)
		}
	}
}

func TestResolveRef(t *testing.T) {
	specPath := testutil.TestSpecPath(t, "petstore-3.0.json")

	spec, err := ParseSpec(specPath)
	if err != nil {
		t.Fatalf("ParseSpec(%q) failed: %v", specPath, err)
	}

	// Resolve a reference to Pet schema
	ref := "#/components/schemas/Pet"
	result, err := ResolveRef(spec, ref)
	if err != nil {
		t.Fatalf("ResolveRef(%q) failed: %v", ref, err)
	}

	// Result should be a *RawSchemaOrRef
	schemaRef, ok := result.(*RawSchemaOrRef)
	if !ok {
		t.Fatalf("ResolveRef(%q) returned type %T, want *RawSchemaOrRef", ref, result)
	}
	if schemaRef.Properties == nil {
		t.Fatal("resolved schema should have properties")
	}
	if _, ok := schemaRef.Properties["id"]; !ok {
		t.Errorf("expected property 'id' in resolved Pet schema")
	}
}

func TestResolveSchemaRef(t *testing.T) {
	specPath := testutil.TestSpecPath(t, "petstore-3.0.json")

	spec, err := ParseSpec(specPath)
	if err != nil {
		t.Fatalf("ParseSpec(%q) failed: %v", specPath, err)
	}

	raw := RawSchemaOrRef{
		Ref: "#/components/schemas/Pet",
	}
	ref, err := ResolveSchemaRef(spec, &raw)
	if err != nil {
		t.Fatalf("ResolveSchemaRef failed: %v", err)
	}
	if ref == nil {
		t.Fatal("ResolveSchemaRef returned nil")
	}
	// Pet has no explicit "type" — ensure properties are resolved
	if ref.Properties == nil {
		t.Fatal("resolved schema should have properties")
	}
	if _, ok := ref.Properties["id"]; !ok {
		t.Errorf("expected property 'id' in resolved schema")
	}
}

func TestHTTPMethods(t *testing.T) {
	expected := []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"}
	if len(HTTPMethods) != len(expected) {
		t.Errorf("HTTPMethods length = %d, want %d", len(HTTPMethods), len(expected))
	}
	for i, m := range HTTPMethods {
		if m != expected[i] {
			t.Errorf("HTTPMethods[%d] = %q, want %q", i, m, expected[i])
		}
	}
}

func TestJSONRoundTrip(t *testing.T) {
	spec := &Spec{
		OpenAPI: "3.0.0",
		Info: map[string]any{
			"title":   "Test",
			"version": "1.0.0",
		},
		Paths: map[string]*PathItem{
			"/test": {
				Get: &Operation{
					OperationID: "testOp",
					Responses: map[string]RawResponseOrRef{
						"200": {
							Response: Response{Description: "OK"},
						},
					},
				},
			},
		},
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Spec
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.OpenAPI != "3.0.0" {
		t.Errorf("decoded.OpenAPI = %q", decoded.OpenAPI)
	}
	if decoded.Info["title"] != "Test" {
		t.Errorf("decoded.Info['title'] = %q", decoded.Info["title"])
	}
	if decoded.Paths["/test"] == nil {
		t.Fatal("expected /test path after round-trip")
	}
	if decoded.Paths["/test"].Get == nil {
		t.Fatal("expected GET /test after round-trip")
	}
	if decoded.Paths["/test"].Get.OperationID != "testOp" {
		t.Errorf("decoded operationId = %q", decoded.Paths["/test"].Get.OperationID)
	}
}
