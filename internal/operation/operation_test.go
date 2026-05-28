package operation

import (
	"testing"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
	"github.com/czemar/ng-openapi-gen/internal/testutil"
)

func testOpts() *config.Options {
	return &config.Options{
		FetchTimeout: 20000,
		DefaultTag:   "Api",
		EnumStyle:    "alias",
	}
}

func TestNewOperationPetstore(t *testing.T) {
	spec, err := openapi.ParseSpec(testutil.TestSpecPath(t, "petstore-3.0.json"))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	opts := testOpts()

	t.Run("list pets", func(t *testing.T) {
		pathSpec := spec.Paths["/pets"]
		op := NewOperation(spec, "/pets", pathSpec, "get", "listPets", pathSpec.Get, opts)
		if op.ID != "listPets" {
			t.Errorf("ID = %q, want %q", op.ID, "listPets")
		}
		if op.Path != "/pets" {
			t.Errorf("Path = %q, want %q", op.Path, "/pets")
		}
		if op.Method != "get" {
			t.Errorf("Method = %q, want %q", op.Method, "get")
		}
		if len(op.Parameters) == 0 {
			t.Errorf("expected at least 1 parameter (limit)")
		}
		if len(op.Variants) == 0 {
			t.Errorf("expected at least 1 variant")
		}
		if len(op.AllResponses) == 0 {
			t.Errorf("expected at least 1 response")
		}
		if len(op.Tags) == 0 {
			t.Errorf("expected tags")
		}
	})

	t.Run("show pet by id has path param", func(t *testing.T) {
		pathSpec := spec.Paths["/pets/{petId}"]
		op := NewOperation(spec, "/pets/{petId}", pathSpec, "get", "showPetById", pathSpec.Get, opts)
		if len(op.Parameters) == 0 {
			t.Errorf("expected path parameter (petId)")
		}
	})

	t.Run("operation without request body is fine", func(t *testing.T) {
		pathSpec := spec.Paths["/pets"]
		op := NewOperation(spec, "/pets", pathSpec, "post", "createPets", pathSpec.Post, opts)
		if op.RequestBody != nil {
			t.Log("createPets has request body (optional in spec)")
		}
	})
}

func TestNewOperationAllOps(t *testing.T) {
	spec, err := openapi.ParseSpec(testutil.TestSpecPath(t, "all-operations.json"))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	opts := testOpts()

	tests := []struct {
		path   string
		method string
		id     string
	}{
		{"/path1", "get", "path1Get"},
		{"/path1", "post", "path1Post"},
		{"/path2/{id}", "get", "path2IdGet"},
		{"/path5", "get", "path5Get"},
		{"/path8", "get", "path8Get"},
		{"/path8", "post", "path8Post"},
	}

	for _, tt := range tests {
		t.Run(tt.path+"."+tt.method, func(t *testing.T) {
			pathSpec := spec.Paths[tt.path]
			opSpec := openapi.GetMethodOperation(pathSpec, tt.method)
			if opSpec == nil {
				t.Fatal("operation not found")
			}
			op := NewOperation(spec, tt.path, pathSpec, tt.method, tt.id, opSpec, opts)
			if op.ID == "" {
				t.Errorf("ID should not be empty")
			}
			if len(op.Variants) == 0 {
				t.Errorf("expected variants")
			}
		})
	}
}

func TestNewOperationVariant(t *testing.T) {
	spec, err := openapi.ParseSpec(testutil.TestSpecPath(t, "petstore-3.0.json"))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	opts := testOpts()
	pathSpec := spec.Paths["/pets"]
	op := NewOperation(spec, "/pets", pathSpec, "get", "listPets", pathSpec.Get, opts)

	t.Run("variant properties", func(t *testing.T) {
		if len(op.Variants) == 0 {
			t.Fatal("no variants")
		}
		v := op.Variants[0]
		if v.MethodName == "" {
			t.Errorf("MethodName should not be empty")
		}
		if v.ResultType == "" {
			t.Errorf("ResultType should not be empty")
		}
	})

	t.Run("variant imports", func(t *testing.T) {
		for _, v := range op.Variants {
			if v.SuccessResponse == nil {
				t.Errorf("variant %s should have success response", v.MethodName)
			}
			if v.ParamsImport == nil {
				t.Errorf("variant %s should have params import", v.MethodName)
			}
		}
	})
}

func TestNewContent(t *testing.T) {
	c := NewContent("application/json", &openapi.MediaType{}, testOpts(), &openapi.Spec{})
	if c == nil {
		t.Fatal("NewContent returned nil")
	}
	if c.MediaType != "application/json" {
		t.Errorf("MediaType = %q, want %q", c.MediaType, "application/json")
	}
}

func TestNewResponse(t *testing.T) {
	spec, err := openapi.ParseSpec(testutil.TestSpecPath(t, "petstore-3.0.json"))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	opts := testOpts()

	t.Run("list pets has responses", func(t *testing.T) {
		pathSpec := spec.Paths["/pets"]
		op := NewOperation(spec, "/pets", pathSpec, "get", "listPets", pathSpec.Get, opts)
		if len(op.AllResponses) == 0 {
			t.Fatal("expected responses")
		}
		for _, resp := range op.AllResponses {
			if resp.StatusCode == "" {
				t.Errorf("StatusCode should not be empty")
			}
		}
	})

	t.Run("NewResponse constructor", func(t *testing.T) {
		content := []*Content{
			NewContent("application/json", &openapi.MediaType{}, opts, spec),
		}
		r := NewResponse("200", "OK", content, opts)
		if r.StatusCode != "200" {
			t.Errorf("StatusCode = %q, want %q", r.StatusCode, "200")
		}
		if r.Description != "OK" {
			t.Errorf("Description = %q, want %q", r.Description, "OK")
		}
		if len(r.Content) != 1 {
			t.Errorf("expected 1 content, got %d", len(r.Content))
		}
	})
}

func TestNewParameter(t *testing.T) {
	spec, err := openapi.ParseSpec(testutil.TestSpecPath(t, "petstore-3.0.json"))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	opts := testOpts()

	pathSpec := spec.Paths["/pets"]
	op := NewOperation(spec, "/pets", pathSpec, "get", "listPets", pathSpec.Get, opts)

	var limitParam *Parameter
	for _, p := range op.Parameters {
		if p.Name == "limit" {
			limitParam = p
			break
		}
	}
	if limitParam == nil {
		t.Fatal("expected limit parameter")
	}
	if limitParam.Name != "limit" {
		t.Errorf("Name = %q, want %q", limitParam.Name, "limit")
	}
	if limitParam.In != "query" {
		t.Errorf("limit should be a query parameter, got %q", limitParam.In)
	}
}

func TestVariantTag(t *testing.T) {
	spec, err := openapi.ParseSpec(testutil.TestSpecPath(t, "all-operations.json"))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	opts := testOpts()

	pathSpec := spec.Paths["/path5"]
	op := NewOperation(spec, "/path5", pathSpec, "get", "path5Get", pathSpec.Get, opts)

	for _, v := range op.Variants {
		tag := v.Tag()
		if tag == "" {
			t.Errorf("Tag() should not be empty")
		}
	}
}


