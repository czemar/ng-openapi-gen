package operation

import (
	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/gen"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

// RequestBody represents an operation request body
type RequestBody struct {
	Spec       *openapi.RequestBody
	Content    []*Content
	TsComments string
	Required   bool
}

// NewRequestBody creates a new RequestBody
func NewRequestBody(spec *openapi.RequestBody, content []*Content, opts *config.Options) *RequestBody {
	rb := &RequestBody{
		Spec:       spec,
		Content:    content,
		TsComments: gen.TsComments(spec.Description, 2),
		Required:   spec.Required,
	}
	return rb
}
