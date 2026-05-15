package operation

import (
	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/gen"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

// Content represents a media type in a request body or response
type Content struct {
	MediaType string
	Spec      *openapi.MediaType
	Type      string // TypeScript type string
}

// NewContent creates a new Content
func NewContent(mediaType string, spec *openapi.MediaType, opts *config.Options, oa *openapi.Spec) *Content {
	c := &Content{
		MediaType: mediaType,
		Spec:      spec,
		Type:      gen.TsType(&spec.Schema, opts, oa, ""),
	}
	return c
}
