package openapi

import (
	"testing"
)

func TestContentOrder(t *testing.T) {
	spec, err := ParseSpec("../../test/all-operations.json")
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range spec.ContentTypeOrder {
		t.Logf("  %s -> %v", k, v)
	}
	
	key := "paths./path3/{id}.delete.responses.200.content"
	if order, ok := spec.ContentTypeOrder[key]; ok {
		t.Logf("Found key %q: %v", key, order)
	} else {
		t.Errorf("key %q not found in ContentTypeOrder", key)
	}
}
