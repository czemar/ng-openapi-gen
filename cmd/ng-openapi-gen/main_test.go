package main

import (
	"testing"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

func TestIsURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"http://example.com/spec.yaml", true},
		{"https://example.com/spec.yaml", true},
		{"https://", true},
		{"/path/to/file.json", false},
		{"file.json", false},
		{"", false},
		{"http", false},
	}
	for _, tt := range tests {
		got := isURL(tt.input)
		if got != tt.want {
			t.Errorf("isURL(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestMakeSet(t *testing.T) {
	tests := []struct {
		input []string
		want  map[string]bool
	}{
		{[]string{"a", "b", "c"}, map[string]bool{"a": true, "b": true, "c": true}},
		{[]string{}, map[string]bool{}},
		{nil, map[string]bool{}},
		{[]string{"dup", "dup"}, map[string]bool{"dup": true}},
	}
	for _, tt := range tests {
		got := makeSet(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("makeSet(%v) length = %d, want %d", tt.input, len(got), len(tt.want))
		}
		for k := range tt.want {
			if !got[k] {
				t.Errorf("makeSet(%v) missing key %q", tt.input, k)
			}
		}
	}
}

func TestGetOperation(t *testing.T) {
	item := &openapi.PathItem{
		Get:    &openapi.Operation{OperationID: "getOp"},
		Post:   &openapi.Operation{OperationID: "postOp"},
		Put:    &openapi.Operation{OperationID: "putOp"},
		Delete: &openapi.Operation{OperationID: "deleteOp"},
	}

	tests := []struct {
		method string
		wantID string
	}{
		{"get", "getOp"},
		{"post", "postOp"},
		{"put", "putOp"},
		{"delete", "deleteOp"},
		{"patch", ""},
		{"unknown", ""},
	}
	for _, tt := range tests {
		op := getOperation(item, tt.method)
		if op == nil && tt.wantID != "" {
			t.Errorf("getOperation(%q) returned nil, want non-nil", tt.method)
			continue
		}
		if op != nil && op.OperationID != tt.wantID {
			t.Errorf("getOperation(%q).OperationID = %q, want %q", tt.method, op.OperationID, tt.wantID)
		}
	}
}

func TestSetOperationNil(t *testing.T) {
	item := &openapi.PathItem{
		Get:  &openapi.Operation{OperationID: "getOp"},
		Post: &openapi.Operation{OperationID: "postOp"},
	}

	setOperationNil(item, "get")
	if item.Get != nil {
		t.Error("Get should be nil after setOperationNil")
	}
	if item.Post == nil {
		t.Error("Post should still be set")
	}

	setOperationNil(item, "patch")
	// Should not panic on methods that don't exist
}

func TestFilterPaths(t *testing.T) {
	// Should not panic on nil paths
	filterPaths(&openapi.Spec{Paths: nil}, nil)

	spec := &openapi.Spec{
		Paths: map[string]*openapi.PathItem{
			"/pets": {
				Get: &openapi.Operation{OperationID: "listPets", Tags: []string{"pets"}},
			},
			"/users": {
				Get: &openapi.Operation{OperationID: "listUsers", Tags: []string{"users"}},
			},
		},
	}
	opts := &config.Options{
		ExcludePaths: []string{"/users"},
	}
	filterPaths(spec, opts)
	if _, ok := spec.Paths["/users"]; ok {
		t.Error("expected /users to be removed")
	}
	if _, ok := spec.Paths["/pets"]; !ok {
		t.Error("expected /pets to remain")
	}
}

func TestFilterPathsExcludeTags(t *testing.T) {
	spec := &openapi.Spec{
		Paths: map[string]*openapi.PathItem{
			"/pets": {
				Get: &openapi.Operation{OperationID: "listPets", Tags: []string{"pets"}},
			},
			"/admin": {
				Get: &openapi.Operation{OperationID: "adminOp", Tags: []string{"admin"}},
			},
		},
	}
	opts := &config.Options{
		ExcludeTags: []string{"admin"},
	}
	filterPaths(spec, opts)
	if _, ok := spec.Paths["/admin"]; ok {
		t.Error("expected /admin to be removed")
	}
	if _, ok := spec.Paths["/pets"]; !ok {
		t.Error("expected /pets to remain")
	}
}

func TestFilterPathsIncludeTags(t *testing.T) {
	spec := &openapi.Spec{
		Paths: map[string]*openapi.PathItem{
			"/pets": {
				Get: &openapi.Operation{OperationID: "listPets", Tags: []string{"pets"}},
			},
			"/admin": {
				Get: &openapi.Operation{OperationID: "adminOp", Tags: []string{"admin"}},
			},
		},
	}
	opts := &config.Options{
		IncludeTags: []string{"pets"},
	}
	filterPaths(spec, opts)
	if _, ok := spec.Paths["/admin"]; ok {
		t.Error("expected /admin to be removed (not in include tags)")
	}
	if _, ok := spec.Paths["/pets"]; !ok {
		t.Error("expected /pets to remain")
	}
}

func TestFilterPathsFullPathRemoval(t *testing.T) {
	spec := &openapi.Spec{
		Paths: map[string]*openapi.PathItem{
			"/pets": {
				Get:  &openapi.Operation{OperationID: "listPets", Tags: []string{"pets"}},
				Post: &openapi.Operation{OperationID: "createPets", Tags: []string{"pets"}},
			},
			"/admin": {
				Get: &openapi.Operation{OperationID: "adminOp", Tags: []string{"admin"}},
			},
		},
	}
	opts := &config.Options{
		ExcludeTags: []string{"pets"},
	}
	filterPaths(spec, opts)
	// All /pets operations are tagged "pets" so the entire path is removed
	if _, ok := spec.Paths["/pets"]; ok {
		t.Error("expected /pets path to be removed")
	}
	if _, ok := spec.Paths["/admin"]; !ok {
		t.Error("expected /admin to remain")
	}
}
