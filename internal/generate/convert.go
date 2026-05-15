package generate

import (
	"strings"

	"github.com/czemar/ng-openapi-gen/internal/gen"
	"github.com/czemar/ng-openapi-gen/internal/model"
	"github.com/czemar/ng-openapi-gen/internal/operation"
	"github.com/czemar/ng-openapi-gen/internal/service"
)

// modelToMap converts a Model to a map for template rendering
func modelToMap(m *model.Model) map[string]any {
	result := map[string]any{
		"typeName":                 m.TypeName,
		"fileName":                 m.FileName,
		"tsComments":               m.TsComments,
		"pathToRoot":               m.PathToRoot,
		"isSimple":                 m.IsSimple,
		"isEnum":                   m.IsEnum,
		"isObject":                 m.IsObject,
		"simpleType":               m.SimpleType,
		"enumArrayName":            m.EnumArrayName,
		"enumArrayFileName":        m.EnumArrayFileName,
		"additionalPropertiesType": m.AdditionalPropertiesType,
		"orphanRequiredProperties": m.OrphanRequiredProperties,
		"namespace":                m.Namespace,
		"qualifiedName":            m.QualifiedName,
		"name":                     m.Name,
	}

	imports := make([]map[string]any, len(m.Imports))
	for i, imp := range m.Imports {
		imports[i] = importToMap(imp)
	}
	result["imports"] = imports

	evs := make([]map[string]any, len(m.EnumValues))
	for i, ev := range m.EnumValues {
		evs[i] = map[string]any{
			"name":        ev.Name,
			"value":       ev.Value,
			"description": ev.Description,
			"type":        ev.Type,
		}
	}
	result["enumValues"] = evs

	props := make([]map[string]any, len(m.Properties))
	for i, prop := range m.Properties {
		props[i] = map[string]any{
			"identifier": prop.Identifier,
			"tsComments": prop.TsComments,
			"type":       prop.Type,
			"name":       prop.Name,
			"required":   prop.Required,
		}
	}
	result["properties"] = props

	return result
}

func variantToMap(v *operation.OperationVariant) map[string]any {
	result := map[string]any{
		"methodName":               v.MethodName,
		"responseMethodName":       v.ResponseMethodName,
		"resultType":               v.ResultType,
		"responseType":             v.ResponseType,
		"accept":                   v.Accept,
		"pathToRoot":               strings.Repeat("../", len(strings.Split(v.ImportPath, "/"))),
		"isVoid":                   v.IsVoid,
		"isNumber":                 v.IsNumber,
		"isBoolean":                v.IsBoolean,
		"isOther":                  v.IsOther,
		"responseMethodTsComments": v.ResponseMethodTsComments,
		"bodyMethodTsComments":     v.BodyMethodTsComments,
		"paramsType":               v.ParamsType,
		"importName":               v.ImportName,
		"importPath":               v.ImportPath,
		"importFile":               v.ImportFile,
		"exportName":               v.ExportName,
		"paramsTypeExportName":     v.ParamsTypeExportName,
	}

	imports := make([]map[string]any, len(v.Imports))
	for i, imp := range v.Imports {
		imports[i] = importToMap(imp)
	}
	result["imports"] = imports

	if v.RequestBody != nil {
		result["requestBody"] = map[string]any{
			"mediaType": v.RequestBody.MediaType,
			"type":      v.RequestBody.Type,
		}
	}

	if v.Operation != nil {
		op := v.Operation
		opMap := map[string]any{
			"path":               op.Path,
			"method":             op.Method,
			"parametersRequired": op.ParametersRequired,
			"pathVar":            op.PathVar,
			"id":                 op.ID,
			"deprecated":         op.Deprecated,
		}

		params := make([]map[string]any, len(op.Parameters))
		for i, p := range op.Parameters {
			params[i] = map[string]any{
				"name":             p.Name,
				"var":              p.Var,
				"varAccess":        p.VarAccess,
				"tsComments":       p.TsComments,
				"required":         p.Required,
				"in":               p.In,
				"type":             p.Type,
				"style":            p.Style,
				"parameterOptions": p.ParameterOptions,
			}
		}
		opMap["parameters"] = params

		if op.RequestBody != nil {
			opMap["requestBody"] = map[string]any{
				"tsComments": op.RequestBody.TsComments,
				"required":   op.RequestBody.Required,
			}
		}

		result["operation"] = opMap
	}

	return result
}

func serviceToMap(s *service.Service) map[string]any {
	result := map[string]any{
		"typeName":   s.TypeName,
		"fileName":   s.FileName,
		"tsComments": s.TsComments,
		"pathToRoot": s.PathToRoot,
	}

	imports := make([]map[string]any, len(s.Imports))
	for i, imp := range s.Imports {
		imports[i] = importToMap(imp)
	}
	result["imports"] = imports

	ops := make([]map[string]any, len(s.Operations))
	for i, op := range s.Operations {
		opMap := map[string]any{
			"path":    op.Path,
			"method":  op.Method,
			"id":      op.ID,
			"pathVar": op.PathVar,
		}
		variants := make([]map[string]any, len(op.Variants))
		for j, v := range op.Variants {
			variants[j] = variantToMap(v)
		}
		opMap["variants"] = variants
		ops[i] = opMap
	}
	result["operations"] = ops

	return result
}

func importToMap(imp *gen.Import) map[string]any {
	if imp == nil {
		return nil
	}
	return map[string]any{
		"name":          imp.Name,
		"typeName":      imp.TypeName,
		"qualifiedName": imp.QualifiedName,
		"path":          imp.Path,
		"file":          imp.File,
		"fullPath":      imp.FullPath,
		"useAlias":      imp.UseAlias,
		"typeOnly":      imp.TypeOnly,
	}
}
