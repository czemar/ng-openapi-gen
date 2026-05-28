package service

import (
	"testing"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
	"github.com/czemar/ng-openapi-gen/internal/operation"
	"github.com/czemar/ng-openapi-gen/internal/testutil"
)

func testOpts() *config.Options {
	return &config.Options{
		FetchTimeout: 20000,
		DefaultTag:   "Api",
		EnumStyle:    "alias",
	}
}

func newTestOperation(t *testing.T, spec *openapi.Spec, path, method, id string) *operation.Operation {
	t.Helper()
	pathSpec := spec.Paths[path]
	if pathSpec == nil {
		t.Fatalf("path %s not found", path)
	}
	opSpec := openapi.GetMethodOperation(pathSpec, method)
	if opSpec == nil {
		t.Fatalf("%s %s not found", method, path)
	}
	return operation.NewOperation(spec, path, pathSpec, method, id, opSpec, testOpts())
}



func TestNewService(t *testing.T) {
	spec, err := openapi.ParseSpec(testutil.TestSpecPath(t, "petstore-3.0.json"))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	opts := testOpts()

	ops := []*operation.Operation{
		newTestOperation(t, spec, "/pets", "get", "listPets"),
		newTestOperation(t, spec, "/pets", "post", "createPets"),
		newTestOperation(t, spec, "/pets/{petId}", "get", "showPetById"),
	}

	svc := NewService("pets", "Operations about pets", ops, opts)
	if svc.TypeName != "PetsService" {
		t.Errorf("TypeName = %q, want %q", svc.TypeName, "PetsService")
	}
	if len(svc.Operations) != 3 {
		t.Errorf("expected 3 operations, got %d", len(svc.Operations))
	}
	if svc.FileName != "pets.service" {
		t.Errorf("FileName = %q, want %q", svc.FileName, "pets.service")
	}
	if svc.TagName != "pets" {
		t.Errorf("TagName = %q, want %q", svc.TagName, "pets")
	}
	if svc.TagDescription != "Operations about pets" {
		t.Errorf("TagDescription = %q, want %q", svc.TagDescription, "Operations about pets")
	}
	if len(svc.Imports) == 0 {
		t.Errorf("expected imports (from operation variants)")
	}
}

func TestNewServiceWithPrefixSuffix(t *testing.T) {
	spec, err := openapi.ParseSpec(testutil.TestSpecPath(t, "petstore-3.0.json"))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	opts := &config.Options{
		ServicePrefix:  "Api",
		ServiceSuffix:  "",
		FetchTimeout:   20000,
		DefaultTag:     "Api",
		EnumStyle:      "alias",
	}

	ops := []*operation.Operation{
		newTestOperation(t, spec, "/pets", "get", "listPets"),
	}

	svc := NewService("pets", "", ops, opts)
	if svc.TypeName != "ApiPetsService" {
		t.Errorf("TypeName = %q, want %q", svc.TypeName, "ApiPetsService")
	}
}

func TestNewServiceDefaultTag(t *testing.T) {
	spec, err := openapi.ParseSpec(testutil.TestSpecPath(t, "petstore-3.0.json"))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	opts := testOpts()
	opts.DefaultTag = "MyApi"

	ops := []*operation.Operation{
		newTestOperation(t, spec, "/pets", "get", "listPets"),
	}

	svc := NewService("pets", "", ops, opts)
	if svc.TypeName != "PetsService" {
		t.Errorf("TypeName = %q, want %q", svc.TypeName, "PetsService")
	}
}

func TestServiceGetImportName(t *testing.T) {
	spec, err := openapi.ParseSpec(testutil.TestSpecPath(t, "petstore-3.0.json"))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	opts := testOpts()
	ops := []*operation.Operation{
		newTestOperation(t, spec, "/pets", "get", "listPets"),
	}
	svc := NewService("pets", "", ops, opts)

	if name := svc.GetImportName(); name != svc.TypeName {
		t.Errorf("GetImportName() = %q, want %q", name, svc.TypeName)
	}
	if path := svc.GetImportPath(); path != "services" {
		t.Errorf("GetImportPath() = %q, want %q", path, "services")
	}
	if f := svc.GetImportFile(); f != svc.FileName {
		t.Errorf("GetImportFile() = %q, want %q", f, svc.FileName)
	}
	if tn := svc.GetImportTypeName(); tn != svc.TypeName {
		t.Errorf("GetImportTypeName() = %q, want %q", tn, svc.TypeName)
	}
	if qn := svc.GetImportQualifiedName(); qn != svc.TypeName {
		t.Errorf("GetImportQualifiedName() = %q, want %q", qn, svc.TypeName)
	}
}
