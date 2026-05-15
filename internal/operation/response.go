package operation

import (
	"github.com/czemar/ng-openapi-gen/internal/config"
)

// Response represents an operation response
type Response struct {
	StatusCode  string
	Description string
	Content     []*Content
	Options     *config.Options
}

// NewResponse creates a new Response
func NewResponse(statusCode, description string, content []*Content, opts *config.Options) *Response {
	return &Response{
		StatusCode:  statusCode,
		Description: description,
		Content:     content,
		Options:     opts,
	}
}
