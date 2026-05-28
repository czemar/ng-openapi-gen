// Package testutil provides shared testing helpers for packages across the
// project, reducing boilerplate duplication in test files.
package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func FindProjectRoot(t testing.TB) string {
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

func TestSpecPath(t testing.TB, name string) string {
	t.Helper()
	return filepath.Join(FindProjectRoot(t), "test", name)
}

func BoolPtr(b bool) *bool {
	return &b
}
