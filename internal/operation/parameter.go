package operation

import (
	"strings"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/gen"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

// Parameter represents an operation parameter
type Parameter struct {
	Var              string
	VarAccess        string
	Name             string
	TsComments       string
	Required         bool
	In               string
	Type             string
	Style            string
	Explode          *bool
	ParameterOptions string
	Specific         bool
	Spec             *openapi.Parameter
}

// NewParameter creates a new Parameter
func NewParameter(spec *openapi.Parameter, opts *config.Options, oa *openapi.Spec) *Parameter {
	p := &Parameter{
		Spec: spec,
		Name: spec.Name,
		In:   spec.In,
	}
	if p.In == "" {
		p.In = "query"
	}
	p.Var = gen.EscapeId(p.Name)
	if strings.Contains(p.Var, "'") {
		p.VarAccess = "[" + p.Var + "]"
	} else {
		p.VarAccess = "." + p.Var
	}
	p.TsComments = gen.TsComments(spec.Description, 0, spec.Deprecated)
	p.Required = spec.In == "path" || spec.Required
	p.Type = gen.TsType(&spec.Schema, opts, oa, "")
	p.Style = spec.Style
	p.Explode = spec.Explode
	p.ParameterOptions = p.createParameterOptions()
	return p
}

func (p *Parameter) createParameterOptions() string {
	var buf strings.Builder
	buf.WriteByte('{')
	first := true
	if p.Style != "" {
		buf.WriteString(`"style":"`)
		buf.WriteString(p.Style)
		buf.WriteByte('"')
		first = false
	}
	if p.Explode != nil {
		if !first {
			buf.WriteByte(',')
		}
		if *p.Explode {
			buf.WriteString(`"explode":true`)
		} else {
			buf.WriteString(`"explode":false`)
		}
	}
	buf.WriteByte('}')
	return buf.String()
}
