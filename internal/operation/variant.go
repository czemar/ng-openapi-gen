package operation

import (
	"strings"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/gen"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

// OperationVariant represents one request/response content type combination
type OperationVariant struct {
	Operation                *Operation
	MethodName               string
	ResponseMethodName       string
	RequestBody              *Content
	SuccessResponse          *Content
	Options                  *config.Options
	ResultType               string
	ResponseType             string
	Accept                   string
	IsVoid                   bool
	IsNumber                 bool
	IsBoolean                bool
	IsOther                  bool
	ResponseMethodTsComments string
	BodyMethodTsComments     string
	ParamsType               string
	ParamsImport             *variantImport
	ImportName               string
	ImportPath               string
	ImportFile               string
	ExportName               string
	ParamsTypeExportName     string
	Imports                  []*gen.Import
	ImportSet                *gen.Imports
}

type variantImport struct {
	ImportName          string
	ImportFile          string
	ImportPath          string
	ImportTypeName      string
	ImportQualifiedName string
}

func (v *variantImport) GetImportName() string          { return v.ImportName }
func (v *variantImport) GetImportPath() string          { return v.ImportPath }
func (v *variantImport) GetImportFile() string          { return v.ImportFile }
func (v *variantImport) GetImportTypeName() string      { return v.ImportTypeName }
func (v *variantImport) GetImportQualifiedName() string { return v.ImportQualifiedName }

// NewOperationVariant creates a new operation variant
func NewOperationVariant(op *Operation, methodName string, reqBody, successResp *Content, opts *config.Options) *OperationVariant {
	v := &OperationVariant{
		Operation:       op,
		MethodName:      methodName,
		RequestBody:     reqBody,
		SuccessResponse: successResp,
		Options:         opts,
		ImportSet:       gen.NewImports(opts, ""),
	}

	v.ResponseMethodName = methodName + "$Response"

	if successResp != nil {
		v.ResultType = successResp.Type
		v.ResponseType = v.inferResponseType(successResp, op)
		v.Accept = successResp.MediaType
	} else {
		v.ResultType = "void"
		v.ResponseType = "text"
		v.Accept = "*/*"
	}

	v.IsVoid = v.ResultType == "void"
	v.IsNumber = v.ResultType == "number"
	v.IsBoolean = v.ResultType == "boolean"
	v.IsOther = !v.IsVoid && !v.IsNumber && !v.IsBoolean

	v.ImportPath = "fn/" + gen.FileName(op.Tag())
	v.ImportName = gen.EnsureNotReserved(methodName)
	v.ImportFile = gen.FileName(methodName)

	v.ParamsType = gen.UpperFirst(methodName) + "$Params"
	v.ParamsImport = &variantImport{
		ImportName: v.ParamsType,
		ImportFile: v.ImportFile,
		ImportPath: v.ImportPath,
	}

	// Collect imports from parameters
	for _, param := range op.Parameters {
		v.ImportSet.AddRef(&param.Spec.Schema, false, op.OpenApi, opts)
	}
	// Collect from request body
	if reqBody != nil && reqBody.Spec != nil {
		v.ImportSet.AddRef(&reqBody.Spec.Schema, false, op.OpenApi, opts)
	}
	// Collect from response
	if successResp != nil && successResp.Spec != nil {
		v.ImportSet.AddRef(&successResp.Spec.Schema, false, op.OpenApi, opts)
	}

	v.updateImports()

	return v
}

func (v *OperationVariant) inferResponseType(successResp *Content, op *Operation) string {
	// Check customized response type by path
	if op.Options.CustomizedResponseType != nil {
		if cr, ok := op.Options.CustomizedResponseType[op.Path]; ok {
			return cr.ToUse
		}
	}

	// Check binary format
	schemaRef := &successResp.Spec.Schema
	schema, err := openapi.ResolveSchemaRef(op.OpenApi, schemaRef)
	if err == nil && schema != nil && schema.Format == "binary" {
		return "blob"
	}

	mediaType := strings.ToLower(successResp.MediaType)
	if strings.Contains(mediaType, "/json") || strings.Contains(mediaType, "+json") {
		return "json"
	} else if strings.HasPrefix(mediaType, "text/") {
		return "text"
	}
	return "blob"
}

func (v *OperationVariant) updateImports() {
	v.Imports = v.ImportSet.ToArray()
	for _, imp := range v.Imports {
		imp.Path = v.pathToRoot() + "models/"
		imp.FullPath = "models/" + imp.File
	}
}

func (v *OperationVariant) pathToRoot() string {
	depth := len(strings.Split(v.ImportPath, "/"))
	return strings.Repeat("../", depth)
}

// Tag returns the operation tag
func (v *OperationVariant) Tag() string {
	if len(v.Operation.Tags) > 0 {
		return v.Operation.Tags[0]
	}
	if v.Options != nil && v.Options.DefaultTag != "" {
		return v.Options.DefaultTag
	}
	return "Api"
}

func (v *OperationVariant) descriptionPrefix() string {
	desc := strings.TrimSpace(v.Operation.Description)
	summary := v.Operation.Summary
	if summary != "" {
		if !strings.HasSuffix(summary, ".") {
			summary += "."
		}
		if desc != "" {
			summary += "\n\n" + desc
		}
		desc = summary
	}
	if desc != "" {
		desc += "\n\n"
	}
	return desc
}

func (v *OperationVariant) descriptionSuffix() string {
	sends := ""
	if v.RequestBody != nil {
		sends = "sends `" + v.RequestBody.MediaType + "` and "
	}
	handles := "doesn't expect any request body"
	if v.RequestBody != nil {
		handles = "handles request body of type `" + v.RequestBody.MediaType + "`"
	}
	return "\n\nThis method " + sends + handles + "."
}

// GetImportName implements Importable
func (v *OperationVariant) GetImportName() string          { return v.ImportName }
func (v *OperationVariant) GetImportPath() string          { return v.ImportPath }
func (v *OperationVariant) GetImportFile() string          { return v.ImportFile }
func (v *OperationVariant) GetImportTypeName() string      { return v.ImportName }
func (v *OperationVariant) GetImportQualifiedName() string { return v.ImportName }
