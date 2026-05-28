// Package openapi parses OpenAPI 3.0 and 3.1 specifications from JSON and YAML
// files into the internal type system used by the code generator.
package openapi

import (
	"bytes"
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

	// Extract content type order from raw JSON
	spec.ContentTypeOrder = extractContentTypeOrder(data)
	return &spec, nil
}

// extractContentTypeOrder walks raw JSON and records content type key order.
func extractContentTypeOrder(data []byte) map[string][]string {
	result := make(map[string][]string)

	var jsonData []byte
	if isJSON(data) {
		jsonData = data
	} else {
		// For YAML, convert to JSON first
		var yamlData any
		if err := yaml.Unmarshal(data, &yamlData); err != nil {
			return result
		}
		var err error
		jsonData, err = json.Marshal(yamlData)
		if err != nil {
			return result
		}
	}

	// Parse into generic structure
	var root map[string]any
	if err := json.Unmarshal(jsonData, &root); err != nil {
		return result
	}

	// Walk paths
	paths, _ := root["paths"].(map[string]any)
	for pathKey := range paths {
		pathItem, _ := paths[pathKey].(map[string]any)
		for _, method := range []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"} {
			op, _ := pathItem[method].(map[string]any)
			if op == nil {
				continue
			}
			base := "paths." + pathKey + "." + method

			// Extract from responses
			if responses, _ := op["responses"].(map[string]any); responses != nil {
				for statusCode := range responses {
					resp, _ := responses[statusCode].(map[string]any)
					if resp == nil {
						continue
					}
					if content, _ := resp["content"].(map[string]any); content != nil {
						contentKey := base + ".responses." + statusCode + ".content"
						keys := orderedKeysFromJSON(jsonData, contentKey)
						if len(keys) > 0 {
							result[contentKey] = keys
						}
					}
				}
			}

			// Extract from requestBody
			if rb, _ := op["requestBody"].(map[string]any); rb != nil {
				if content, _ := rb["content"].(map[string]any); content != nil {
					contentKey := base + ".requestBody.content"
					keys := orderedKeysFromJSON(jsonData, contentKey)
					if len(keys) > 0 {
						result[contentKey] = keys
					}
				}
			}
		}
	}

	return result
}

// orderedKeysFromJSON extracts the ordered keys of a nested object from raw JSON.
// The path is dot-separated, e.g. "paths./path3/{id}.delete.responses.200.content"
func orderedKeysFromJSON(data []byte, path string) []string {
	parts := strings.Split(path, ".")
	dec := json.NewDecoder(bytes.NewReader(data))

	// Skip the opening { of the root object
	firstTok, err := dec.Token()
	if err != nil {
		return nil
	}
	if _, ok := firstTok.(json.Delim); !ok {
		return nil
	}

	return findOrderedKeys(dec, parts)
}

// findOrderedKeys recursively walks JSON following path parts, collecting keys at the final part.
func findOrderedKeys(dec *json.Decoder, parts []string) []string {
	if len(parts) == 0 {
		return nil
	}
	part := parts[0]
	isLast := len(parts) == 1

	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil
		}
		key, ok := tok.(string)
		if !ok {
			return nil
		}

		// Read the value
		valueTok, err := dec.Token()
		if err != nil {
			return nil
		}

		if key == part {
			if isLast {
				// Collect keys from the value object
				if delim, ok := valueTok.(json.Delim); ok && delim == '{' {
					return collectObjectKeys(dec)
				}
				return nil
			}
			// Descend into this object
			if delim, ok := valueTok.(json.Delim); ok && delim == '{' {
				return findOrderedKeys(dec, parts[1:])
			}
			return nil
		}

		// Skip non-matching values
		skipValue(dec, valueTok)
	}

	// Consume closing delimiter (shouldn't reach here for valid JSON)
	dec.Token()
	return nil
}

// collectObjectKeys reads all keys from a JSON object (assumes opening { already consumed)
func collectObjectKeys(dec *json.Decoder) []string {
	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return keys
		}
		key, ok := tok.(string)
		if !ok {
			return keys
		}
		keys = append(keys, key)

		// Skip the value
		valueTok, err := dec.Token()
		if err != nil {
			return keys
		}
		skipValue(dec, valueTok)
	}
	dec.Token() // consume closing }
	return keys
}

// skipValue skips a JSON value, handling nested objects/arrays
func skipValue(dec *json.Decoder, tok json.Token) {
	if delim, ok := tok.(json.Delim); ok {
		switch delim {
		case '{':
			for dec.More() {
				// Skip key
				dec.Token()
				// Skip value
				valTok, _ := dec.Token()
				skipValue(dec, valTok)
			}
			dec.Token() // consume closing }
		case '[':
			for dec.More() {
				elemTok, _ := dec.Token()
				skipValue(dec, elemTok)
			}
			dec.Token() // consume closing ]
		}
	}
}

func isJSON(data []byte) bool {
	data = bytes.TrimSpace(data)
	return len(data) > 0 && data[0] == '{'
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


