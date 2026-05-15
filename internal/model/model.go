package model

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/gen"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

// Model represents a generated model type
type Model struct {
	OpenApi    *openapi.Spec   `json:"-"`
	Name       string          `json:"name"`
	Schema     *openapi.Schema `json:"-"`
	Options    *config.Options `json:"-"`
	TypeName   string          `json:"typeName"`
	FileName   string          `json:"fileName"`
	TsComments string          `json:"tsComments"`
	PathToRoot string          `json:"pathToRoot"`
	Imports    []*gen.Import   `json:"imports"`
	ImportSet  *gen.Imports    `json:"-"`

	IsSimple bool `json:"isSimple"`
	IsEnum   bool `json:"isEnum"`
	IsObject bool `json:"isObject"`

	SimpleType string `json:"simpleType"`

	EnumValues        []*EnumValue `json:"enumValues"`
	EnumArrayName     string       `json:"enumArrayName"`
	EnumArrayFileName string       `json:"enumArrayFileName"`

	Properties               []*Property `json:"properties"`
	AdditionalPropertiesType string      `json:"additionalPropertiesType"`
	OrphanRequiredProperties []string    `json:"orphanRequiredProperties"`

	Namespace     string `json:"namespace"`
	QualifiedName string `json:"qualifiedName"`

	additionalPropertiesSchema *openapi.Schema `json:"-"`
}

// NewModel creates a new Model from a schema
func NewModel(oa *openapi.Spec, name string, schema *openapi.Schema, opts *config.Options) *Model {
	m := &Model{
		OpenApi:    oa,
		Name:       name,
		Schema:     schema,
		Options:    opts,
		TypeName:   gen.UnqualifiedName(name, opts),
		Namespace:  gen.Namespace(name),
		ImportSet:  gen.NewImports(opts, gen.UnqualifiedName(name, opts)),
		Properties: nil,
	}

	m.FileName = gen.FileName(m.TypeName)
	m.QualifiedName = m.TypeName
	if m.Namespace != "" {
		m.FileName = m.Namespace + "/" + m.FileName
		m.QualifiedName = gen.TypeName(m.Namespace, opts) + m.TypeName
	}

	m.TsComments = gen.TsComments(schema.Description, 0, schema.Deprecated)

	schemaType := resolveSchemaTypeSimple(schema)
	typeForEnum := schemaType
	if types, ok := schemaType.([]string); ok && len(types) > 0 {
		typeForEnum = types[0]
	}

	typeStr, _ := typeForEnum.(string)
	if len(schema.Enum) > 0 && typeStr != "" && (typeStr == "string" || typeStr == "number" || typeStr == "integer") {
		m.EnumArrayName = strings.ToUpper(strings.ReplaceAll(m.TypeName, " ", "_"))
		m.EnumArrayFileName = gen.FileName(m.TypeName + "-array")

		names := schema.XEnumNames
		descriptions := schema.XEnumDescriptions
		if names == nil {
			names = []string{}
		}
		if descriptions == nil {
			descriptions = []string{}
		}

		m.EnumValues = make([]*EnumValue, len(schema.Enum))
		for i, val := range schema.Enum {
			evName := ""
			evDesc := ""
			if i < len(names) {
				evName = names[i]
			}
			if i < len(descriptions) {
				evDesc = descriptions[i]
			}
			m.EnumValues[i] = NewEnumValue(typeStr, evName, evDesc, val, opts)
		}

		if opts.EnumStyle != "alias" {
			m.IsEnum = true
		}
	}

	hasAllOf := len(schema.AllOf) > 0
	hasOneOf := len(schema.OneOf) > 0
	m.IsObject = (typeStr == "object" || len(schema.Properties) > 0) && !openapi.IsNullableSchema(schema) && !hasAllOf && !hasOneOf
	m.IsSimple = !m.IsObject && !m.IsEnum

	if m.IsObject {
		propsByName := make(map[string]*Property)
		m.collectObject(schema, propsByName)
		sortedNames := make([]string, 0, len(propsByName))
		for name := range propsByName {
			sortedNames = append(sortedNames, name)
		}
		sort.Strings(sortedNames)
		m.Properties = make([]*Property, len(sortedNames))
		for i, name := range sortedNames {
			m.Properties[i] = propsByName[name]
		}
	}

	// Collect imports
	m.collectImportsFromSchema(schema)
	m.updateImports()

	if hasAllOf {
		m.collectOrphanRequiredProperties(schema)
	}

	if m.IsObject {
		for _, prop := range m.Properties {
			prop.ResolveType(opts, oa, m.TypeName)
		}
		m.finalizeAdditionalPropertiesType()
	}

	if m.IsSimple {
		m.SimpleType = gen.TsType(&openapi.RawSchemaOrRef{Schema: *schema}, opts, oa, m.TypeName)
	}

	return m
}

func (m *Model) collectObject(schema *openapi.Schema, propsByName map[string]*Property) {
	if schema.Type == "object" || len(schema.Properties) > 0 {
		required := makeSet(schema.Required)
		propNames := make([]string, 0, len(schema.Properties))
		for name := range schema.Properties {
			propNames = append(propNames, name)
		}
		sort.Strings(propNames)

		for _, name := range propNames {
			prop := schema.Properties[name]
			propCopy := prop
			if _, exists := propsByName[name]; !exists {
				propsByName[name] = NewProperty(m.TypeName, name, propCopy, required[name], m.Options, m.OpenApi)
			}
		}

		if add, ok := schema.AdditionalProperties.(bool); ok && add {
			m.AdditionalPropertiesType = "any"
		} else if addMap, ok := schema.AdditionalProperties.(map[string]any); ok {
			m.additionalPropertiesSchema = mapToSchema(addMap)
		}
	}

	for _, item := range schema.AllOf {
		item := item
		if item.Ref != "" {
			resolved, err := openapi.ResolveSchemaRef(m.OpenApi, &item)
			if err == nil && resolved != nil {
				m.collectObject(resolved, propsByName)
			}
		} else {
			m.collectObject(&item.Schema, propsByName)
		}
	}
}

func (m *Model) finalizeAdditionalPropertiesType() {
	if m.additionalPropertiesSchema == nil {
		return
	}

	propTypes := make(map[string]bool)
	appendType := func(typeStr string) {
		if strings.HasPrefix(typeStr, "null | ") {
			propTypes["null"] = true
			propTypes[strings.TrimPrefix(typeStr, "null | ")] = true
		} else {
			propTypes[typeStr] = true
		}
	}

	for _, prop := range m.Properties {
		appendType(prop.Type)
		if !prop.Required {
			propTypes["undefined"] = true
		}
	}

	addType := gen.TsType(&openapi.RawSchemaOrRef{Schema: *m.additionalPropertiesSchema}, m.Options, m.OpenApi, m.TypeName)
	appendType(addType)

	types := make([]string, 0, len(propTypes))
	for t := range propTypes {
		types = append(types, t)
	}
	sort.Strings(types)
	m.AdditionalPropertiesType = strings.Join(types, " | ")
}

func (m *Model) collectOrphanRequiredProperties(schema *openapi.Schema) {
	for _, subschema := range schema.AllOf {
		subschema := subschema
		if subschema.Ref != "" {
			continue
		}
		if len(subschema.Schema.Required) > 0 && len(subschema.Schema.Properties) == 0 {
			m.OrphanRequiredProperties = append(m.OrphanRequiredProperties, subschema.Schema.Required...)
		}
	}
}

func (m *Model) collectImportsFromSchema(schema *openapi.Schema) {
	for _, item := range schema.OneOf {
		item := item
		m.ImportSet.AddRef(&item, false, m.OpenApi, m.Options)
	}
	for _, item := range schema.AllOf {
		item := item
		m.ImportSet.AddRef(&item, false, m.OpenApi, m.Options)
	}
	for _, item := range schema.AnyOf {
		item := item
		m.ImportSet.AddRef(&item, false, m.OpenApi, m.Options)
	}
	for _, item := range schema.PrefixItems {
		item := item
		m.ImportSet.AddRef(&item, false, m.OpenApi, m.Options)
	}
	if gen.IsArraySchema(schema) && schema.Items != nil {
		m.ImportSet.AddRef(schema.Items, false, m.OpenApi, m.Options)
	}
	propNames := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		propNames = append(propNames, name)
	}
	sort.Strings(propNames)
	for _, name := range propNames {
		prop := schema.Properties[name]
		m.ImportSet.AddRef(&prop, false, m.OpenApi, m.Options)
	}
	if addMap, ok := schema.AdditionalProperties.(map[string]any); ok {
		if ref, ok := addMap["$ref"]; ok {
			if refStr, ok := ref.(string); ok {
				m.ImportSet.Add(gen.SimpleName(refStr), false)
			}
		}
	}
}

func (m *Model) updateImports() {
	m.PathToRoot = "../"
	if m.Namespace != "" {
		depth := strings.Count(m.Namespace, "/") + 1
		m.PathToRoot = strings.Repeat("../", depth)
	}
	m.Imports = m.ImportSet.ToArray()
	for _, imp := range m.Imports {
		imp.Path = m.PathToRoot + "models/"
		imp.FullPath = "models/" + imp.File
	}
}

func resolveSchemaTypeSimple(schema *openapi.Schema) any {
	if schema.Type != nil {
		return schema.Type
	}
	return "object"
}

func makeSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, item := range items {
		m[item] = true
	}
	return m
}

func mapToSchema(m map[string]any) *openapi.Schema {
	data, _ := json.Marshal(m)
	var s openapi.Schema
	json.Unmarshal(data, &s)
	return &s
}
