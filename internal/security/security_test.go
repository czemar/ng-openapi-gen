package security

import (
	"testing"

	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

func TestNewSecurityAPIKey(t *testing.T) {
	spec := &openapi.SecurityScheme{
		Type:        "apiKey",
		Name:        "X-API-Key",
		In:          "header",
		Description: "API key authentication",
	}
	s := NewSecurity("apiKey", spec, nil)

	if s.Var != "apiKey" {
		t.Errorf("Var = %q, want %q", s.Var, "apiKey")
	}
	if s.Name != "X-API-Key" {
		t.Errorf("Name = %q, want %q", s.Name, "X-API-Key")
	}
	if s.In != "header" {
		t.Errorf("In = %q, want %q", s.In, "header")
	}
	if s.Type != "string" {
		t.Errorf("Type = %q, want %q", s.Type, "string")
	}
	if s.TsComments == "" {
		t.Error("TsComments should not be empty")
	}
}

func TestNewSecurityAPIKeyDefaultIn(t *testing.T) {
	spec := &openapi.SecurityScheme{
		Type: "apiKey",
		Name: "token",
		In:   "",
	}
	s := NewSecurity("myAuth", spec, nil)

	if s.In != "header" {
		t.Errorf("In = %q, want 'header' (default)", s.In)
	}
}

func TestNewSecurityNonAPIKey(t *testing.T) {
	spec := &openapi.SecurityScheme{
		Type:        "http",
		Scheme:      "bearer",
		Description: "Bearer token",
	}
	s := NewSecurity("bearer", spec, []string{"read", "write"})

	if s.Name != "" {
		t.Errorf("Name should be empty for non-apiKey, got %q", s.Name)
	}
	if s.Var != "bearer" {
		t.Errorf("Var = %q, want %q", s.Var, "bearer")
	}
	if len(s.Scope) != 2 {
		t.Errorf("Scope length = %d, want 2", len(s.Scope))
	}
}

func TestNewSecurityOAuth2(t *testing.T) {
	spec := &openapi.SecurityScheme{
		Type: "oauth2",
	}
	s := NewSecurity("oauth", spec, []string{"profile"})

	if s.Var != "oauth" {
		t.Errorf("Var = %q, want %q", s.Var, "oauth")
	}
	if s.Type != "string" {
		t.Errorf("Type = %q, want %q", s.Type, "string")
	}
	if len(s.Scope) != 1 || s.Scope[0] != "profile" {
		t.Errorf("Scope = %v, want [profile]", s.Scope)
	}
}

func TestNewSecurityOpenID(t *testing.T) {
	spec := &openapi.SecurityScheme{
		Type:             "openIdConnect",
		OpenIDConnectURL: "https://example.com/.well-known/openid-configuration",
	}
	s := NewSecurity("openid", spec, nil)

	if s.Var != "openid" {
		t.Errorf("Var = %q, want %q", s.Var, "openid")
	}
}
