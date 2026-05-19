package service

import (
	"strings"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/gen"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
	"github.com/czemar/ng-openapi-gen/internal/operation"
)

// Service groups operations by tag
type Service struct {
	TagName                string
	TagDescription         string
	Operations             []*operation.Operation
	Options                *config.Options
	TypeName               string
	FileName               string
	TsComments             string
	PathToRoot             string
	Imports                []*gen.Import
	ImportSet              *gen.Imports
	AdditionalDependencies []string
}

// NewService creates a new service from a tag and its operations
func NewService(tagName, tagDesc string, ops []*operation.Operation, opts *config.Options) *Service {
	s := &Service{
		TagName:        tagName,
		TagDescription: tagDesc,
		Operations:     ops,
		Options:        opts,
		TypeName:       gen.ServiceClass(tagName, opts),
		ImportSet:      gen.NewImports(opts, ""),
	}

	s.FileName = gen.FileName(s.TypeName)
	// Angular standards: xxx.service.ts, not xxx-service.ts
	if strings.HasSuffix(s.FileName, "-service") {
		s.FileName = strings.TrimSuffix(s.FileName, "-service") + ".service"
	}

	s.TsComments = gen.TsComments(tagDesc, 0)

	// Collect imports from all operations
	for _, op := range ops {
		for _, variant := range op.Variants {
			// Import the variant function
			s.ImportSet.Add(variant, false)
			// Import the variant params
			s.ImportSet.Add(variant.ParamsImport, false)

			// Collect response type imports
			if variant.SuccessResponse != nil && variant.SuccessResponse.Spec != nil {
				s.collectImport(&variant.SuccessResponse.Spec.Schema, false)
			}
			// Collect request body imports
			if variant.RequestBody != nil && variant.RequestBody.Spec != nil {
				s.collectImport(&variant.RequestBody.Spec.Schema, true)
			}
		}

		// Parameter imports
		for _, param := range op.Parameters {
			s.collectImport(&param.Spec.Schema, true)
		}

		// Response imports
		for _, resp := range op.AllResponses {
			for _, content := range resp.Content {
				if content.Spec != nil {
					s.collectImport(&content.Spec.Schema, true)
				}
			}
		}
	}

	s.updateImports()

	return s
}

func (s *Service) collectImport(sRaw *openapi.RawSchemaOrRef, additional bool) {
	if sRaw == nil {
		return
	}
	if sRaw.Ref != "" {
		name := gen.SimpleName(sRaw.Ref)
		s.ImportSet.Add(name, additional)
	} else {
		s.collectFromSchema(&sRaw.Schema, additional)
	}
}

func (s *Service) collectFromSchema(schema *openapi.Schema, additional bool) {
	for _, item := range schema.OneOf {
		item := item
		s.collectImport(&item, additional)
	}
	for _, item := range schema.AllOf {
		item := item
		s.collectImport(&item, additional)
	}
	for _, item := range schema.AnyOf {
		item := item
		s.collectImport(&item, additional)
	}
	if gen.IsArraySchema(schema) && schema.Items != nil {
		s.collectImport(schema.Items, additional)
	}
	for _, prop := range schema.Properties {
		prop := prop
		s.collectImport(&prop, additional)
	}
}

func (s *Service) updateImports() {
	s.PathToRoot = "../"
	s.Imports = s.ImportSet.ToArray()
}

// GetImportName implements Importable
func (s *Service) GetImportName() string          { return s.TypeName }
func (s *Service) GetImportPath() string          { return "services" }
func (s *Service) GetImportFile() string          { return s.FileName }
func (s *Service) GetImportTypeName() string      { return s.TypeName }
func (s *Service) GetImportQualifiedName() string { return s.TypeName }
