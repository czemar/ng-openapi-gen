// Package security models security scheme requirements from OpenAPI specs,
// supporting apiKey, http, oauth2, and openIdConnect schemes.
package security

import (
	"github.com/czemar/ng-openapi-gen/internal/gen"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

// Security represents an operation security requirement
type Security struct {
	Var        string
	Name       string
	TsComments string
	In         string
	Type       string
	Spec       *openapi.SecurityScheme
	Scope      []string
}

// NewSecurity creates a new Security from a security scheme
func NewSecurity(key string, spec *openapi.SecurityScheme, scope []string) *Security {
	s := &Security{
		Spec:  spec,
		Scope: scope,
		In:    "header",
		Type:  "string",
	}
	if spec.Type == "apiKey" {
		s.Name = spec.Name
		s.In = spec.In
		if s.In == "" {
			s.In = "header"
		}
	}
	s.Var = gen.MethodName(key)
	s.TsComments = gen.TsComments(spec.Description, 2)

	return s
}
