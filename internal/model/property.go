package model

import (
	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/gen"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

// Property represents an object property
type Property struct {
	Identifier string                  `json:"identifier"`
	TsComments string                  `json:"tsComments"`
	Type       string                  `json:"type"`
	Name       string                  `json:"name"`
	Required   bool                    `json:"required"`
	Spec       *openapi.RawSchemaOrRef `json:"-"`
}

// NewProperty creates a new deferred property (type resolved later)
func NewProperty(containerName string, name string, schema openapi.RawSchemaOrRef, required bool, opts *config.Options, spec *openapi.Spec) *Property {
	p := &Property{
		Name:     name,
		Spec:     &schema,
		Required: required,
		Type:     "",
	}

	propSchema, _ := openapi.ResolveSchemaRef(spec, &schema)
	p.TsComments = gen.TsComments("", 1)
	if propSchema != nil {
		p.TsComments = gen.TsComments(propSchema.Description, 1, propSchema.Deprecated)
	}
	p.Identifier = gen.EscapeId(name)
	return p
}

// ResolveType resolves the TypeScript type after imports are finalized
func (p *Property) ResolveType(opts *config.Options, spec *openapi.Spec, containerName string) {
	p.Type = gen.TsType(p.Spec, opts, spec, containerName)
}
