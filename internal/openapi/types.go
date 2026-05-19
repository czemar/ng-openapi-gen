package openapi

// Raw OpenAPI 3.0/3.1 type definitions for JSON parsing

type Spec struct {
	OpenAPI    string                `json:"openapi"`
	Info       map[string]any        `json:"info,omitempty"`
	Servers    []Server              `json:"servers,omitempty"`
	Paths      map[string]*PathItem  `json:"paths,omitempty"`
	Components *Components           `json:"components,omitempty"`
	Tags       []Tag                 `json:"tags,omitempty"`
	Security   []map[string][]string `json:"security,omitempty"`
	Webhooks   map[string]any        `json:"webhooks,omitempty"`

	// ContentTypeOrder tracks spec order of content type keys
	// key: "paths./path.method.responses.status.content" or "paths./path.method.requestBody.content"
	// e.g. "paths./path3/{id}.delete.responses.200.content"
	ContentTypeOrder map[string][]string
}

type Server struct {
	URL         string            `json:"url"`
	Description string            `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

type ServerVariable struct {
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type PathItem struct {
	Summary     string                         `json:"summary,omitempty"`
	Description string                         `json:"description,omitempty"`
	Get         *Operation                     `json:"get,omitempty"`
	Put         *Operation                     `json:"put,omitempty"`
	Post        *Operation                     `json:"post,omitempty"`
	Delete      *Operation                     `json:"delete,omitempty"`
	Options     *Operation                     `json:"options,omitempty"`
	Head        *Operation                     `json:"head,omitempty"`
	Patch       *Operation                     `json:"patch,omitempty"`
	Trace       *Operation                     `json:"trace,omitempty"`
	Parameters  []RawParameterOrRef            `json:"parameters,omitempty"`
}

type Operation struct {
	Tags         []string                     `json:"tags,omitempty"`
	Summary      string                       `json:"summary,omitempty"`
	Description  string                       `json:"description,omitempty"`
	OperationID  string                       `json:"operationId,omitempty"`
	Deprecated   bool                         `json:"deprecated,omitempty"`
	Parameters   []RawParameterOrRef          `json:"parameters,omitempty"`
	RequestBody  *RawRequestBodyOrRef         `json:"requestBody,omitempty"`
	Responses    map[string]RawResponseOrRef  `json:"responses,omitempty"`
	Security     *[]map[string][]string       `json:"security,omitempty"`
	Servers      []Server                     `json:"servers,omitempty"`
	XOperationName string                     `json:"x-operation-name,omitempty"`
}

type Components struct {
	Schemas         map[string]RawSchemaOrRef `json:"schemas,omitempty"`
	Responses       map[string]RawResponseOrRef `json:"responses,omitempty"`
	Parameters      map[string]RawParameterOrRef `json:"parameters,omitempty"`
	RequestBodies   map[string]RawRequestBodyOrRef `json:"requestBodies,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`
}

type Tag struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	ExternalDocs any    `json:"externalDocs,omitempty"`
}

// Reference object
type Ref struct {
	Ref string `json:"$ref"`
}

// RawParameterOrRef is either a Parameter or a $ref
type RawParameterOrRef struct {
	Ref    string     `json:"$ref,omitempty"`
	Parameter
}

// RawRequestBodyOrRef is either a RequestBody or a $ref
type RawRequestBodyOrRef struct {
	Ref    string     `json:"$ref,omitempty"`
	RequestBody
}

// RawResponseOrRef is either a Response or a $ref
type RawResponseOrRef struct {
	Ref    string   `json:"$ref,omitempty"`
	Response
}

// RawSchemaOrRef is either a Schema or a $ref
type RawSchemaOrRef struct {
	Ref    string  `json:"$ref,omitempty"`
	Schema
}

type Parameter struct {
	Name            string            `json:"name"`
	In              string            `json:"in"`
	Description     string            `json:"description,omitempty"`
	Required        bool              `json:"required,omitempty"`
	Deprecated      bool              `json:"deprecated,omitempty"`
	AllowEmptyValue bool              `json:"allowEmptyValue,omitempty"`
	Style           string            `json:"style,omitempty"`
	Explode         *bool             `json:"explode,omitempty"`
	AllowReserved   bool              `json:"allowReserved,omitempty"`
	Schema          RawSchemaOrRef    `json:"schema,omitempty"`
	Example         any               `json:"example,omitempty"`
	Examples        map[string]any    `json:"examples,omitempty"`
	Content         map[string]MediaType `json:"content,omitempty"`
}

type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

type Response struct {
	Description string               `json:"description"`
	Headers     map[string]any       `json:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

type MediaType struct {
	Schema   RawSchemaOrRef  `json:"schema,omitempty"`
	Example  any             `json:"example,omitempty"`
	Examples map[string]any  `json:"examples,omitempty"`
	Encoding map[string]any  `json:"encoding,omitempty"`
}

type SecurityScheme struct {
	Type             string              `json:"type"`
	Description      string              `json:"description,omitempty"`
	Name             string              `json:"name,omitempty"`
	In               string              `json:"in,omitempty"`
	Scheme           string              `json:"scheme,omitempty"`
	BearerFormat     string              `json:"bearerFormat,omitempty"`
	Flows            any                 `json:"flows,omitempty"`
	OpenIDConnectURL string              `json:"openIdConnectUrl,omitempty"`
}

type Schema struct {
	Ref                  string                    `json:"$ref,omitempty"`
	Type                 any                       `json:"type,omitempty"` // string or []string in 3.1
	Format               string                    `json:"format,omitempty"`
	Title                string                    `json:"title,omitempty"`
	Description          string                    `json:"description,omitempty"`
	Deprecated           bool                      `json:"deprecated,omitempty"`
	Nullable             bool                      `json:"nullable,omitempty"` // OpenAPI 3.0
	Required             []string                  `json:"required,omitempty"`
	Enum                 []any                     `json:"enum,omitempty"`
	Const                any                       `json:"const,omitempty"`
	Items                *RawSchemaOrRef           `json:"items,omitempty"`
	PrefixItems          []RawSchemaOrRef          `json:"prefixItems,omitempty"` // OpenAPI 3.1 tuples
	AllOf                []RawSchemaOrRef          `json:"allOf,omitempty"`
	OneOf                []RawSchemaOrRef          `json:"oneOf,omitempty"`
	AnyOf                []RawSchemaOrRef          `json:"anyOf,omitempty"`
	Not                  *RawSchemaOrRef           `json:"not,omitempty"`
	Properties           map[string]RawSchemaOrRef `json:"properties,omitempty"`
	AdditionalProperties any                       `json:"additionalProperties,omitempty"` // bool or Schema
	Discriminator        *Discriminator            `json:"discriminator,omitempty"`
	ReadOnly             bool                      `json:"readOnly,omitempty"`
	WriteOnly            bool                      `json:"writeOnly,omitempty"`
	XML                  map[string]any            `json:"xml,omitempty"`
	ExternalDocs         any                       `json:"externalDocs,omitempty"`
	Example              any                       `json:"example,omitempty"`
	Default              any                       `json:"default,omitempty"`
	Minimum              *float64                  `json:"minimum,omitempty"`
	Maximum              *float64                  `json:"maximum,omitempty"`
	ExclusiveMinimum     any                       `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     any                       `json:"exclusiveMaximum,omitempty"`
	MinLength            *int                      `json:"minLength,omitempty"`
	MaxLength            *int                      `json:"maxLength,omitempty"`
	Pattern              string                    `json:"pattern,omitempty"`
	MinItems             *int                      `json:"minItems,omitempty"`
	MaxItems             *int                      `json:"maxItems,omitempty"`
	UniqueItems          bool                      `json:"uniqueItems,omitempty"`
	XEnumNames           []string                  `json:"x-enumNames,omitempty"`
	XEnumDescriptions    []string                  `json:"x-enumDescriptions,omitempty"`
}

type Discriminator struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty"`
}

var HTTPMethods = []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"}

func (s *Schema) GetType() any {
	if s.Type != nil {
		return s.Type
	}
	return nil
}

// IsRef returns true if this is a $ref reference
func (r RawSchemaOrRef) IsRef() bool {
	return r.Ref != ""
}

func (r RawParameterOrRef) IsRef() bool {
	return r.Ref != ""
}

func (r RawRequestBodyOrRef) IsRef() bool {
	return r.Ref != ""
}

func (r RawResponseOrRef) IsRef() bool {
	return r.Ref != ""
}
