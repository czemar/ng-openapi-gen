package openapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseSpec reads and parses an OpenAPI spec file (JSON or YAML)
func ParseSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(path))
	var spec Spec
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("parse YAML: %w", err)
		}
	default:
		if err := json.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("parse JSON: %w", err)
		}
	}
	return &spec, nil
}

// ResolveRef resolves a $ref string against the spec
func ResolveRef(spec *Spec, ref string) (any, error) {
	if !strings.HasPrefix(ref, "#/") {
		return nil, fmt.Errorf("external refs not supported: %s", ref)
	}

	parts := strings.Split(strings.TrimPrefix(ref, "#/"), "/")
	current := any(spec)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if current == nil {
			return nil, fmt.Errorf("cannot resolve ref %s: nil at %s", ref, part)
		}

		switch v := current.(type) {
		case map[string]any:
			current = v[part]
		case *Spec:
			switch part {
			case "openapi": current = v.OpenAPI
			case "info": current = v.Info
			case "servers": current = v.Servers
			case "paths": current = v.Paths
			case "components": current = v.Components
			case "tags": current = v.Tags
			case "security": current = v.Security
			case "webhooks": current = v.Webhooks
			default: current = nil
			}
		case *Components:
			current = resolveComponentsField(v, part)
		case Components:
			current = resolveComponentsField(&v, part)
		case map[string]RawSchemaOrRef:
			if s, ok := v[part]; ok {
				current = &s
			} else {
				current = nil
			}
		case map[string]*PathItem:
			if p, ok := v[part]; ok {
				current = p
			} else {
				current = nil
			}
		case map[string]RawResponseOrRef:
			if r, ok := v[part]; ok {
				current = &r
			} else {
				current = nil
			}
		case map[string]RawParameterOrRef:
			if p, ok := v[part]; ok {
				current = &p
			} else {
				current = nil
			}
		case map[string]RawRequestBodyOrRef:
			if r, ok := v[part]; ok {
				current = &r
			} else {
				current = nil
			}
		case *SecurityScheme:
			current = v
		case map[string]*SecurityScheme:
			if s, ok := v[part]; ok {
				current = s
			} else {
				current = nil
			}
		case *PathItem:
			switch part {
			case "summary": current = v.Summary
			case "description": current = v.Description
			case "get": current = v.Get
			case "put": current = v.Put
			case "post": current = v.Post
			case "delete": current = v.Delete
			case "options": current = v.Options
			case "head": current = v.Head
			case "patch": current = v.Patch
			case "trace": current = v.Trace
			case "parameters": current = v.Parameters
			default: current = nil
			}
		case *Operation:
			switch part {
			case "tags": current = v.Tags
			case "summary": current = v.Summary
			case "description": current = v.Description
			case "operationId": current = v.OperationID
			case "deprecated": current = v.Deprecated
			case "parameters": current = v.Parameters
			case "requestBody": current = v.RequestBody
			case "responses": current = v.Responses
			case "security": current = v.Security
			case "servers": current = v.Servers
			default: current = nil
			}
		case *RawSchemaOrRef:
			current = &v.Schema
		case *Schema:
			switch part {
			case "$ref": current = v.Ref
			case "type": current = v.Type
			case "format": current = v.Format
			case "title": current = v.Title
			case "description": current = v.Description
			case "deprecated": current = v.Deprecated
			case "nullable": current = v.Nullable
			case "required": current = v.Required
			case "enum": current = v.Enum
			case "const": current = v.Const
			case "items": current = v.Items
			case "prefixItems": current = v.PrefixItems
			case "allOf": current = v.AllOf
			case "oneOf": current = v.OneOf
			case "anyOf": current = v.AnyOf
			case "not": current = v.Not
			case "properties": current = v.Properties
			case "additionalProperties": current = v.AdditionalProperties
			case "discriminator": current = v.Discriminator
			case "readOnly": current = v.ReadOnly
			case "writeOnly": current = v.WriteOnly
			case "xml": current = v.XML
			case "externalDocs": current = v.ExternalDocs
			case "example": current = v.Example
			case "default": current = v.Default
			default: current = nil
			}
		case *RawResponseOrRef:
			current = &v.Response
		case *Response:
			switch part {
			case "description": current = v.Description
			case "headers": current = v.Headers
			case "content": current = v.Content
			default: current = nil
			}
		case *RawRequestBodyOrRef:
			current = &v.RequestBody
		case *RequestBody:
			switch part {
			case "description": current = v.Description
			case "required": current = v.Required
			case "content": current = v.Content
			default: current = nil
			}
		case *RawParameterOrRef:
			current = &v.Parameter
		case *Parameter:
			switch part {
			case "name": current = v.Name
			case "in": current = v.In
			case "description": current = v.Description
			case "required": current = v.Required
			case "deprecated": current = v.Deprecated
			case "allowEmptyValue": current = v.AllowEmptyValue
			case "style": current = v.Style
			case "explode": current = v.Explode
			case "allowReserved": current = v.AllowReserved
			case "schema": current = &v.Schema
			case "example": current = v.Example
			case "examples": current = v.Examples
			case "content": current = v.Content
			default: current = nil
			}
		case map[string]MediaType:
			if mt, ok := v[part]; ok {
				current = &mt
			} else {
				current = nil
			}
		case *MediaType:
			switch part {
			case "schema": current = &v.Schema
			case "example": current = v.Example
			case "examples": current = v.Examples
			case "encoding": current = v.Encoding
			default: current = nil
			}
		case []Tag:
			current = nil
		default:
			return nil, fmt.Errorf("cannot resolve ref %s: unexpected type %T at %s", ref, current, part)
		}
	}
	return current, nil
}

func resolveComponentsField(c *Components, field string) any {
	switch field {
	case "schemas":
		return c.Schemas
	case "responses":
		return c.Responses
	case "parameters":
		return c.Parameters
	case "requestBodies":
		return c.RequestBodies
	case "securitySchemes":
		return c.SecuritySchemes
	}
	return nil
}

// ResolveSchemaRef resolves a schema reference and returns the Schema
func ResolveSchemaRef(spec *Spec, sRaw *RawSchemaOrRef) (*Schema, error) {
	if sRaw == nil {
		return nil, nil
	}
	if sRaw.Ref != "" {
		resolved, err := ResolveRef(spec, sRaw.Ref)
		if err != nil {
			return nil, err
		}
		switch v := resolved.(type) {
		case *Schema:
			return v, nil
		case **RawSchemaOrRef:
			return &(*v).Schema, nil
		case *RawSchemaOrRef:
			return &v.Schema, nil
		default:
			return nil, fmt.Errorf("ref %s resolved to unexpected type %T", sRaw.Ref, resolved)
		}
	}
	return &sRaw.Schema, nil
}

// ResolveParameterRef resolves a parameter reference
func ResolveParameterRef(spec *Spec, pRaw *RawParameterOrRef) (*Parameter, error) {
	if pRaw == nil {
		return nil, nil
	}
	if pRaw.Ref != "" {
		resolved, err := ResolveRef(spec, pRaw.Ref)
		if err != nil {
			return nil, err
		}
		switch v := resolved.(type) {
		case *Parameter:
			return v, nil
		default:
			return nil, fmt.Errorf("ref %s resolved to unexpected type %T", pRaw.Ref, resolved)
		}
	}
	return &pRaw.Parameter, nil
}

// ResolveResponseRef resolves a response reference
func ResolveResponseRef(spec *Spec, rRaw *RawResponseOrRef) (*Response, error) {
	if rRaw == nil {
		return nil, nil
	}
	if rRaw.Ref != "" {
		resolved, err := ResolveRef(spec, rRaw.Ref)
		if err != nil {
			return nil, err
		}
		switch v := resolved.(type) {
		case *Response:
			return v, nil
		default:
			return nil, fmt.Errorf("ref %s resolved to unexpected type %T", rRaw.Ref, resolved)
		}
	}
	return &rRaw.Response, nil
}

// ResolveRequestBodyRef resolves a request body reference
func ResolveRequestBodyRef(spec *Spec, rbRaw *RawRequestBodyOrRef) (*RequestBody, error) {
	if rbRaw == nil {
		return nil, nil
	}
	if rbRaw.Ref != "" {
		resolved, err := ResolveRef(spec, rbRaw.Ref)
		if err != nil {
			return nil, err
		}
		switch v := resolved.(type) {
		case *RequestBody:
			return v, nil
		default:
			return nil, fmt.Errorf("ref %s resolved to unexpected type %T", rbRaw.Ref, resolved)
		}
	}
	return &rbRaw.RequestBody, nil
}

// SchemaRefName returns the simple name from a schema ref
func SchemaRefName(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}


