package config

import (
	"testing"
)

func TestDefaults(t *testing.T) {
	opts := &Options{}
	opts.setDefaults()

	if opts.Output != "src/app/api" {
		t.Errorf("default output = %q, want %q", opts.Output, "src/app/api")
	}
	if opts.FetchTimeout != 20000 {
		t.Errorf("default fetchTimeout = %d, want %d", opts.FetchTimeout, 20000)
	}
	if opts.DefaultTag != "Api" {
		t.Errorf("default defaultTag = %q, want %q", opts.DefaultTag, "Api")
	}
	if opts.EnumStyle != "alias" {
		t.Errorf("default enumStyle = %q, want %q", opts.EnumStyle, "alias")
	}
	if opts.RemoveStaleFiles == nil || !*opts.RemoveStaleFiles {
		t.Error("default removeStaleFiles should be true")
	}
	if opts.ApiService == nil {
		t.Error("default apiService should not be nil")
	}
}

func TestGetStringOrBool(t *testing.T) {
	tests := []struct {
		name   string
		val    any
		defVal string
		want   string
	}{
		{"nil", nil, "def", ""},
		{"string value", "hello", "def", "hello"},
		{"empty string", "", "def", "def"},
		{"true bool", true, "MyService", "MyService"},
		{"false bool", false, "MyService", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStringOrBool(tt.val, tt.defVal)
			if got != tt.want {
				t.Errorf("GetStringOrBool(%v, %q) = %q, want %q", tt.val, tt.defVal, got, tt.want)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want bool
	}{
		{"nil", nil, false},
		{"true bool", true, true},
		{"false bool", false, false},
		{"non-bool", "hello", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBool(tt.val)
			if got != tt.want {
				t.Errorf("GetBool(%v) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

func TestKebabToCamel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello-world", "helloWorld"},
		{"foo-bar-baz", "fooBarBaz"},
		{"single", "single"},
		{"", ""},
		{"alreadyCamel", "alreadyCamel"},
	}
	for _, tt := range tests {
		got := kebabToCamel(tt.input)
		if got != tt.want {
			t.Errorf("kebabToCamel(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
