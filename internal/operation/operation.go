package operation

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/gen"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
	"github.com/czemar/ng-openapi-gen/internal/security"
)

// Operation represents an API endpoint operation
type Operation struct {
	OpenApi  *openapi.Spec
	Path     string
	PathSpec *openapi.PathItem
	Method   string
	ID       string
	Spec     *openapi.Operation
	Options  *config.Options

	Tags               []string
	MethodName         string
	PathVar            string
	Parameters         []*Parameter
	HasParameters      bool
	ParametersRequired bool
	SecurityReqs       [][]*security.Security
	Deprecated         bool
	RequestBody        *RequestBody
	SuccessResponse    *Response
	AllResponses       []*Response
	PathExpression     string
	Variants           []*OperationVariant
	Description        string
	Summary            string
}

// NewOperation creates a new Operation from an OpenAPI operation spec
func NewOperation(oa *openapi.Spec, path string, pathSpec *openapi.PathItem, method, id string, spec *openapi.Operation, opts *config.Options) *Operation {
	op := &Operation{
		OpenApi:     oa,
		Path:        strings.ReplaceAll(path, "'", "\\'"),
		PathSpec:    pathSpec,
		Method:      method,
		ID:          id,
		Spec:        spec,
		Options:     opts,
		Tags:        spec.Tags,
		PathVar:     gen.UpperFirst(id) + "Path",
		Deprecated:  spec.Deprecated,
		Description: spec.Description,
		Summary:     spec.Summary,
	}

	// x-operation-name vendor extension
	if spec.XOperationName != "" {
		op.MethodName = spec.XOperationName
	} else {
		op.MethodName = id
	}

	// Collect parameters
	allParams := op.collectParameters(false, pathSpec.Parameters)
	allParams = append(allParams, op.collectParameters(true, spec.Parameters)...)

	// Deduplicate: specific params override shared params
	op.Parameters = make([]*Parameter, 0, len(allParams))
	for _, param := range allParams {
		skip := false
		if !param.Specific {
			for _, p := range allParams {
				if p != param && p.Name == param.Name && p.Specific {
					skip = true
					break
				}
			}
		}
		if !skip {
			op.Parameters = append(op.Parameters, param)
		}
	}

	for _, p := range op.Parameters {
		if p.Required {
			op.ParametersRequired = true
			break
		}
	}
	op.HasParameters = len(op.Parameters) > 0

	// Security
	if spec.Security != nil {
		op.SecurityReqs = op.collectSecurity(*spec.Security)
	} else if oa.Security != nil {
		op.SecurityReqs = op.collectSecurity(oa.Security)
	}

	// Request body
	if spec.RequestBody != nil && (spec.RequestBody.Ref != "" || spec.RequestBody.Description != "" || len(spec.RequestBody.Content) > 0) {
		rb, err := openapi.ResolveRequestBodyRef(oa, spec.RequestBody)
		if err == nil && rb != nil {
			content := op.collectContent(rb.Content)
			op.RequestBody = NewRequestBody(rb, content, opts)
			if rb.Required {
				op.ParametersRequired = true
			}
		}
	}

	// Responses
	resp := op.collectResponses()
	op.SuccessResponse = resp.success
	op.AllResponses = resp.all

	// Path expression
	op.PathExpression = op.toPathExpression()

	// Calculate variants
	op.calculateVariants()

	return op
}

func (op *Operation) collectParameters(specific bool, params []openapi.RawParameterOrRef) []*Parameter {
	var result []*Parameter
	for _, p := range params {
		p := p
		param, err := openapi.ResolveParameterRef(op.OpenApi, &p)
		if err != nil {
			continue
		}
		if param.In == "cookie" {
			// Skip cookie params
			continue
		}
		if op.paramIsExcluded(param) {
			continue
		}
		parameter := NewParameter(param, op.Options, op.OpenApi)
		parameter.Specific = specific
		result = append(result, parameter)
	}
	return result
}

func (op *Operation) paramIsExcluded(param *openapi.Parameter) bool {
	for _, excluded := range op.Options.ExcludeParameters {
		if excluded == param.Name {
			return true
		}
	}
	return false
}

func (op *Operation) collectSecurity(secReqs []map[string][]string) [][]*security.Security {
	var result [][]*security.Security
	for _, req := range secReqs {
		var group []*security.Security
		for key, scope := range req {
			// Resolve security scheme
			scheme := op.resolveSecurityScheme(key)
			if scheme != nil {
				group = append(group, security.NewSecurity(key, scheme, scope))
			}
		}
		if len(group) > 0 {
			result = append(result, group)
		}
	}
	return result
}

func (op *Operation) resolveSecurityScheme(name string) *openapi.SecurityScheme {
	if op.OpenApi.Components == nil || op.OpenApi.Components.SecuritySchemes == nil {
		return nil
	}
	if scheme, ok := op.OpenApi.Components.SecuritySchemes[name]; ok {
		return scheme
	}
	return nil
}

func (op *Operation) collectContent(content map[string]openapi.MediaType) []*Content {
	var result []*Content
	// Sort keys for deterministic output
	keys := make([]string, 0, len(content))
	for key := range content {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, mediaType := range keys {
		mt := content[mediaType]
		mtCopy := mt
		result = append(result, NewContent(mediaType, &mtCopy, op.Options, op.OpenApi))
	}
	return result
}

type responseResult struct {
	success *Response
	all     []*Response
}

func (op *Operation) collectResponses() responseResult {
	var result responseResult
	responseByType := make(map[string]*Response)

	// Sort response codes for deterministic output
	codes := make([]string, 0, len(op.Spec.Responses))
	for code := range op.Spec.Responses {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	for _, statusCode := range codes {
		resp := op.Spec.Responses[statusCode]
		respCopy := resp
		response := op.getResponse(&respCopy, statusCode)
		result.all = append(result.all, response)

		statusInt := 0
		if n, err := fmt.Sscanf(statusCode, "%d", &statusInt); err == nil && n == 1 {
			if statusInt >= 200 && statusInt < 300 {
				if _, exists := responseByType["successResponse"]; !exists {
					responseByType["successResponse"] = response
				}
			}
		} else if statusCode == "default" {
			responseByType["defaultResponse"] = response
		}
	}

	if sr, ok := responseByType["successResponse"]; ok {
		result.success = sr
	} else if dr, ok := responseByType["defaultResponse"]; ok {
		result.success = dr
	}

	return result
}

func (op *Operation) getResponse(rRaw *openapi.RawResponseOrRef, statusCode string) *Response {
	resp, err := openapi.ResolveResponseRef(op.OpenApi, rRaw)
	if err != nil || resp == nil {
		return NewResponse(statusCode, "", nil, op.Options)
	}
	content := op.collectContent(resp.Content)
	return NewResponse(statusCode, resp.Description, content, op.Options)
}

func (op *Operation) toPathExpression() string {
	re := regexp.MustCompile(`\{([^}]+)}`)
	return re.ReplaceAllStringFunc(op.Path, func(match string) string {
		pName := match[1 : len(match)-1]
		var param *Parameter
		for _, p := range op.Parameters {
			if p.Name == pName {
				param = p
				break
			}
		}
		paramName := pName
		if param != nil {
			paramName = param.Var
		}
		return "${params." + paramName + "}"
	})
}

func (op *Operation) calculateVariants() {
	requestVariants := op.contentsByMethodPart(op.RequestBody)
	responseVariants := op.contentsByMethodPart(op.SuccessResponse)

	totalVariants := max(1, len(requestVariants)) * max(1, len(responseVariants))

	// Check ambiguity
	hasAmbiguity := totalVariants > 1 &&
		len(requestVariants) == 1 && len(responseVariants) == 1

	if hasAmbiguity {
		var reqPart string
		var reqContent *Content
		for k, v := range requestVariants {
			reqPart = k
			reqContent = v
		}
		var respPart string
		var respContent *Content
		for k, v := range responseVariants {
			respPart = k
			respContent = v
		}

		if reqPart == "" && respPart == "" && reqContent != nil && respContent != nil {
			requestMethodPart := op.variantMethodPart(reqContent)
			responseMethodPart := op.variantMethodPart(respContent)

			if requestMethodPart == responseMethodPart {
				// Recalculate with preserved suffixes
				requestVariants = make(map[string]*Content)
				responseVariants = make(map[string]*Content)
				if op.RequestBody != nil {
					for _, content := range op.RequestBody.Content {
						part := op.variantMethodPart(content)
						requestVariants[part] = content
					}
				} else {
					requestVariants[""] = nil
				}
				if op.SuccessResponse != nil {
					for _, content := range op.SuccessResponse.Content {
						part := op.variantMethodPart(content)
						responseVariants[part] = content
					}
				} else {
					responseVariants[""] = nil
				}
			}
		}
	}

	for reqPart, reqContent := range requestVariants {
		for respPart, respContent := range responseVariants {
			methodName := op.MethodName + reqPart + respPart
			v := NewOperationVariant(op, methodName, reqContent, respContent, op.Options)
			op.Variants = append(op.Variants, v)
		}
	}
}

func (op *Operation) contentsByMethodPart(hasContent interface{}) map[string]*Content {
	result := make(map[string]*Content)
	if hasContent != nil {
		switch v := hasContent.(type) {
		case *RequestBody:
			if v != nil && len(v.Content) > 0 {
				for _, content := range v.Content {
					part := op.variantMethodPart(content)
					result[part] = content
				}
			}
		case *Response:
			if v != nil && len(v.Content) > 0 {
				for _, content := range v.Content {
					part := op.variantMethodPart(content)
					result[part] = content
				}
			}
		}
	}

	if len(result) == 0 {
		result[""] = nil
	} else if len(result) == 1 {
		var single *Content
		for _, v := range result {
			single = v
		}
		result = make(map[string]*Content)
		result[""] = single
	}
	return result
}

func (op *Operation) variantMethodPart(content *Content) string {
	if content == nil {
		return ""
	}

	keep := op.keepFullResponseMediaType(content.MediaType)
	mediaType := content.MediaType
	mediaType = strings.ReplaceAll(mediaType, "/*", "")

	if mediaType == "*" || mediaType == "application/octet-stream" {
		return "$Any"
	}

	if keep != "full" {
		parts := strings.Split(mediaType, "/")
		mediaType = parts[len(parts)-1]

		if keep != "tail" {
			plusIdx := strings.LastIndex(mediaType, "+")
			if plusIdx >= 0 {
				mediaType = mediaType[plusIdx+1:]
			}
		}
	}

	if op.Options.SkipJsonSuffix && mediaType == "json" {
		return ""
	}
	return "$" + gen.TypeName(mediaType, op.Options)
}

func (op *Operation) keepFullResponseMediaType(mediaType string) string {
	if op.Options.KeepFullResponseMediaType == true {
		return "full"
	}

	if arr, ok := op.Options.KeepFullResponseMediaType.([]any); ok {
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				use, _ := m["use"].(string)
				if use == "" {
					use = "short"
				}
				mtPattern, _ := m["mediaType"].(string)
				if mtPattern == "" {
					return use
				}
				matched, _ := regexp.MatchString(mtPattern, mediaType)
				if matched {
					return use
				}
			}
		}
	}

	return "short"
}

// Tag returns the first tag or default
func (op *Operation) Tag() string {
	if len(op.Tags) > 0 {
		return op.Tags[0]
	}
	if op.Options.DefaultTag != "" {
		return op.Options.DefaultTag
	}
	return "Api"
}
