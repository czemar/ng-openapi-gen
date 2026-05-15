package gen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

// Import represents a TypeScript import
type Import struct {
	Name          string `json:"name"`
	TypeName      string `json:"typeName"`
	QualifiedName string `json:"qualifiedName"`
	Path          string `json:"path"`
	File          string `json:"file"`
	FullPath      string `json:"fullPath"`
	UseAlias      bool   `json:"useAlias"`
	TypeOnly      bool   `json:"typeOnly"`
}

// Importable represents an artifact that can be imported
type Importable interface {
	GetImportName() string
	GetImportPath() string
	GetImportFile() string
	GetImportTypeName() string
	GetImportQualifiedName() string
}

// Imports manages import statements for a generated file
type Imports struct {
	items           map[string]*Import
	options         *config.Options
	currentTypeName string
}

// NewImports creates a new import manager
func NewImports(opts *config.Options, currentTypeName string) *Imports {
	return &Imports{
		items:           make(map[string]*Import),
		options:         opts,
		currentTypeName: currentTypeName,
	}
}

// Add adds an import, handling collision avoidance
func (m *Imports) Add(param any, typeOnly bool) {
	switch v := param.(type) {
	case string:
		m.addModel(v, typeOnly)
	case Importable:
		m.addImportable(v, typeOnly)
	case *Import:
		if v != nil {
			key := v.Name
			if _, exists := m.items[key]; !exists {
				m.items[key] = v
			}
		}
	}
}

// AddRef adds an import from a raw schema or ref
func (m *Imports) AddRef(sRaw *openapi.RawSchemaOrRef, typeOnly bool, spec *openapi.Spec, opts *config.Options) {
	if sRaw == nil {
		return
	}
	if sRaw.Ref != "" {
		name := SimpleName(sRaw.Ref)
		if m.currentTypeName != name {
			m.Add(name, typeOnly)
		}
	} else {
		m.collectFromSchema(&sRaw.Schema, typeOnly, spec, opts)
	}
}

func (m *Imports) collectFromSchema(schema *openapi.Schema, typeOnly bool, spec *openapi.Spec, opts *config.Options) {
	for _, item := range schema.OneOf {
		item := item
		m.AddRef(&item, typeOnly, spec, opts)
	}
	for _, item := range schema.AllOf {
		item := item
		m.AddRef(&item, typeOnly, spec, opts)
	}
	for _, item := range schema.AnyOf {
		item := item
		m.AddRef(&item, typeOnly, spec, opts)
	}
	for _, item := range schema.PrefixItems {
		item := item
		m.AddRef(&item, typeOnly, spec, opts)
	}
	if IsArraySchema(schema) && schema.Items != nil {
		m.AddRef(schema.Items, typeOnly, spec, opts)
	}
	propNames := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		propNames = append(propNames, name)
	}
	sort.Strings(propNames)
	for _, name := range propNames {
		prop := schema.Properties[name]
		m.AddRef(&prop, typeOnly, spec, opts)
	}
	if addMap, ok := schema.AdditionalProperties.(map[string]any); ok {
		// Simplified - check if it looks like a schema with $ref
		if ref, ok := addMap["$ref"]; ok {
			if refStr, ok := ref.(string); ok {
				name := SimpleName(refStr)
				if m.currentTypeName != name {
					m.Add(name, typeOnly)
				}
			}
		}
	}
}

func (m *Imports) addModel(param string, typeOnly bool) {
	importTypeName := UnqualifiedName(param, m.options)
	importQualifiedName := QualifiedName(param, m.options)

	// Check collision with current type name
	if m.currentTypeName != "" && importTypeName == m.currentTypeName {
		suffix := 1
		aliased := fmt.Sprintf("%s_%d", importTypeName, suffix)
		for m.hasTypeName(aliased) {
			suffix++
			aliased = fmt.Sprintf("%s_%d", importTypeName, suffix)
		}
		importQualifiedName = aliased
	}

	file := FileName(importTypeName)
	if ns := Namespace(param); ns != "" {
		file = ns + "/" + file
	}

	imp := &Import{
		Name:          param,
		TypeName:      importTypeName,
		QualifiedName: importQualifiedName,
		UseAlias:      importTypeName != importQualifiedName,
		TypeOnly:      typeOnly,
		Path:          "models/",
		File:          file,
	}
	imp.FullPath = imp.Path + imp.File
	m.items[imp.Name] = imp
}

func (m *Imports) addImportable(param Importable, typeOnly bool) {
	imp := &Import{
		Name:          param.GetImportName(),
		TypeName:      param.GetImportTypeName(),
		QualifiedName: param.GetImportQualifiedName(),
		UseAlias:      false,
		TypeOnly:      typeOnly,
		Path:          param.GetImportPath() + "/",
		File:          param.GetImportFile(),
	}
	if imp.TypeName == "" {
		imp.TypeName = imp.Name
	}
	if imp.QualifiedName == "" {
		imp.QualifiedName = imp.Name
	}
	imp.UseAlias = imp.TypeName != imp.QualifiedName
	imp.FullPath = strings.Trim(imp.Path+imp.File, "/")
	m.items[imp.Name] = imp
}

func (m *Imports) hasTypeName(typeName string) bool {
	for _, imp := range m.items {
		if imp.QualifiedName == typeName {
			return true
		}
	}
	return false
}

// ToArray returns sorted imports array
func (m *Imports) ToArray() []*Import {
	result := make([]*Import, 0, len(m.items))
	for _, imp := range m.items {
		result = append(result, imp)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.Compare(strings.ToLower(result[i].Name), strings.ToLower(result[j].Name)) < 0
	})
	return result
}

// Size returns the number of imports
func (m *Imports) Size() int {
	return len(m.items)
}
