package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/gen"
	"github.com/czemar/ng-openapi-gen/internal/model"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
	"github.com/czemar/ng-openapi-gen/internal/operation"
	"github.com/czemar/ng-openapi-gen/internal/service"
)

// toMapSlice converts a slice of structs to slice of maps
func toMapSlice[T any](items []T) []map[string]any {
	result := make([]map[string]any, len(items))
	for i, item := range items {
		result[i] = toMap(item)
	}
	return result
}

// toServiceMapSlice converts a map of services to a sorted slice of service maps
func toServiceMapSlice(services map[string]*service.Service) []map[string]any {
	names := make([]string, 0, len(services))
	for name := range services {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]map[string]any, 0, len(services))
	for _, name := range names {
		result = append(result, serviceToMap(services[name]))
	}
	return result
}

// toModelMapSlice converts a map of models to a sorted slice of model maps
func toModelMapSlice(models map[string]*model.Model) []map[string]any {
	names := make([]string, 0, len(models))
	for name := range models {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]map[string]any, 0, len(models))
	for _, name := range names {
		result = append(result, modelToMap(models[name]))
	}
	return result
}

// Logger provides logging functionality
type Logger struct {
	Silent bool
}

func NewLogger(silent bool) *Logger {
	return &Logger{Silent: silent}
}

func (l *Logger) Info(msg string, args ...any) {
	if !l.Silent {
		if len(args) > 0 {
			fmt.Println(fmt.Sprintf(msg, args...))
		} else {
			fmt.Println(msg)
		}
	}
}

func (l *Logger) Debug(msg string, args ...any) {
	if !l.Silent {
		if len(args) > 0 {
			fmt.Println(fmt.Sprintf(msg, args...))
		} else {
			fmt.Println(msg)
		}
	}
}

func (l *Logger) Warn(msg string, args ...any) {
	if len(args) > 0 {
		fmt.Println("[WARN] " + fmt.Sprintf(msg, args...))
	} else {
		fmt.Println("[WARN] " + msg)
	}
}

// Globals stores global generation configuration
type Globals struct {
	ConfigurationClass  string
	ConfigurationFile   string
	ConfigurationParams string
	BaseServiceClass    string
	BaseServiceFile     string
	ApiServiceClass     string
	ApiServiceFile      string
	RequestBuilderClass string
	RequestBuilderFile  string
	ResponseClass       string
	ResponseFile        string
	ModuleClass         string
	ModuleFile          string
	ModelIndexFile      string
	FunctionIndexFile   string
	ServiceIndexFile    string
	RootURL             string
	Promises            bool
	GenerateServices    bool
}

// NewGlobals creates a new Globals from options
func NewGlobals(opts *config.Options) *Globals {
	g := &Globals{
		ConfigurationClass:  getOrDefault(opts.Configuration, "ApiConfiguration"),
		ConfigurationFile:   gen.FileName(getOrDefault(opts.Configuration, "ApiConfiguration")),
		BaseServiceClass:    getOrDefault(opts.BaseService, "BaseService"),
		BaseServiceFile:     gen.FileName(getOrDefault(opts.BaseService, "BaseService")),
		RequestBuilderClass: getOrDefault(opts.RequestBuilder, "RequestBuilder"),
		RequestBuilderFile:  gen.FileName(getOrDefault(opts.RequestBuilder, "RequestBuilder")),
		ResponseClass:       getOrDefault(opts.Response, "StrictHttpResponse"),
		ResponseFile:        gen.FileName(getOrDefault(opts.Response, "StrictHttpResponse")),
		Promises:            opts.Promises == nil || *opts.Promises,
		GenerateServices:    opts.Services != nil && *opts.Services,
	}
	g.ConfigurationParams = g.ConfigurationClass + "Params"

	// API Service
	if opts.ApiService != nil && opts.ApiService != false {
		if s, ok := opts.ApiService.(string); ok && s != "" {
			g.ApiServiceClass = s
		} else {
			g.ApiServiceClass = "Api"
		}
	}
	g.ApiServiceFile = strings.ReplaceAll(gen.FileName(g.ApiServiceClass), "-service", ".service")

	// Module
	if opts.Module != nil && opts.Module != false {
		if s, ok := opts.Module.(string); ok && s != "" {
			g.ModuleClass = s
		} else {
			g.ModuleClass = "ApiModule"
		}
		g.ModuleFile = strings.ReplaceAll(gen.FileName(g.ModuleClass), "-module", ".module")
	}

	// Indexes — mirror TypeScript behavior: default to "models"/"functions"/"services"
	// when the option is nil (not set) or true, skip when explicitly false or empty string
	if opts.ModelIndex == nil || opts.ModelIndex != false {
		if s, ok := opts.ModelIndex.(string); ok && s != "" {
			g.ModelIndexFile = s
		} else {
			g.ModelIndexFile = "models"
		}
	}
	if opts.FunctionIndex == nil || opts.FunctionIndex != false {
		if s, ok := opts.FunctionIndex.(string); ok && s != "" {
			g.FunctionIndexFile = s
		} else {
			g.FunctionIndexFile = "functions"
		}
	}
	if opts.ServiceIndex == nil || opts.ServiceIndex != false {
		if s, ok := opts.ServiceIndex.(string); ok && s != "" {
			g.ServiceIndexFile = s
		} else {
			g.ServiceIndexFile = "services"
		}
	}

	return g
}

// Generator is the main code generation orchestrator
type Generator struct {
	Spec      *openapi.Spec
	Options   *config.Options
	Globals   *Globals
	Logger    *Logger
	OutDir    string
	TempDir   string
	Models    map[string]*model.Model
	Services  map[string]*service.Service
	TemplateM *gen.TemplateManager
}

// NewGenerator creates a new Generator
func NewGenerator(spec *openapi.Spec, opts *config.Options) *Generator {
	g := &Generator{
		Spec:     spec,
		Options:  opts,
		Logger:   NewLogger(opts.Silent),
		OutDir:   strings.TrimRight(opts.Output, "/\\"),
		TempDir:  strings.TrimRight(opts.Output, "/\\") + "$",
		Models:   make(map[string]*model.Model),
		Services: make(map[string]*service.Service),
	}
	g.Globals = NewGlobals(opts)

	// Load templates
	tm, err := gen.NewTemplateManager(opts.Templates)
	if err != nil {
		g.Logger.Warn("Failed to load templates: %v", err)
	}
	g.TemplateM = tm

	// Use temp dir if configured
	if opts.UseTempDir {
		g.TempDir = filepath.Join(os.TempDir(), fmt.Sprintf("ng-openapi-gen-%s$", filepath.Base(g.OutDir)))
	}

	// Validate version
	g.validateVersion()

	// Read root URL
	g.Globals.RootURL = g.readRootURL()

	// Read models
	g.readModels()

	// Read services
	g.readServices()

	// Filter unused models
	if opts.IgnoreUnusedModels == nil || *opts.IgnoreUnusedModels {
		g.ignoreUnusedModels()
	}

	return g
}

// Generate executes code generation
func (g *Generator) Generate() error {
	// Clean and create temp directory
	os.RemoveAll(g.TempDir)
	if err := os.MkdirAll(g.TempDir, 0755); err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(g.TempDir)

	g.Logger.Info("Generating %d models and %d services...", len(g.Models), len(g.Services))

	// Write models
	modelNames := make([]string, 0, len(g.Models))
	for name := range g.Models {
		modelNames = append(modelNames, name)
	}
	sort.Strings(modelNames)
	for _, name := range modelNames {
		m := g.Models[name]
		if err := g.writeTemplate("model", m, m.FileName, "models"); err != nil {
			return err
		}
		if m.EnumArrayFileName != "" {
			if err := g.writeTemplate("enumArray", m, m.EnumArrayFileName, "models"); err != nil {
				return err
			}
		}
	}

	// Collect all functions
	svcNames := make([]string, 0, len(g.Services))
	for name := range g.Services {
		svcNames = append(svcNames, name)
	}
	sort.Strings(svcNames)

	var allFunctions []*operation.OperationVariant
	for _, name := range svcNames {
		svc := g.Services[name]
		for _, op := range svc.Operations {
			for _, variant := range op.Variants {
				allFunctions = append(allFunctions, variant)
			}
		}
	}

	// Write functions
	generateServices := g.Globals.GenerateServices
	for _, name := range svcNames {
		svc := g.Services[name]
		if generateServices {
			if err := g.writeTemplate("service", svc, svc.FileName, "services"); err != nil {
				return err
			}
		}
	}

	// Deduplicate function export names
	// (allFunctions is already in deterministic order: sorted services/paths/methods)
	methodNameCounts := make(map[string]int)
	for _, fn := range allFunctions {
		methodNameCounts[fn.MethodName]++
	}

	for _, fn := range allFunctions {
		if methodNameCounts[fn.MethodName] > 1 {
			tagSuffix := gen.UpperFirst(fn.Tag())
			fn.ExportName = fn.MethodName + tagSuffix
			fn.ParamsTypeExportName = strings.TrimSuffix(fn.ParamsType, "$Params") + tagSuffix + "$Params"
		} else {
			fn.ExportName = fn.ImportName
			fn.ParamsTypeExportName = fn.ParamsType
		}
		if err := g.writeTemplate("fn", fn, fn.ImportFile, fn.ImportPath); err != nil {
			return err
		}
	}

	// Build context for general templates (convert to maps for template access)
	modelSlice := sortedModels(g.Models)
	ctx := map[string]any{
		"services":   toServiceMapSlice(g.Services),
		"models":     toModelMapSlice(g.Models),
		"functions":  toMapSlice(allFunctions),
		"globals":    g.Globals,
		"modelIndex": toMap(newModelIndex(modelSlice, g.Options)),
	}

	// Merge globals into context
	ctx = g.mergeGlobals(ctx)

	// Write general files
	files := []struct {
		template string
		baseName string
		subDir   string
		skip     bool
	}{
		{"configuration", g.Globals.ConfigurationFile, "", false},
		{"response", g.Globals.ResponseFile, "", false},
		{"requestBuilder", g.Globals.RequestBuilderFile, "", false},
		{"baseService", g.Globals.BaseServiceFile, "", !generateServices},
		{"apiService", g.Globals.ApiServiceFile, "", g.Globals.ApiServiceFile == ""},
		{"module", g.Globals.ModuleFile, "", !generateServices || g.Globals.ModuleClass == "" || g.Globals.ModuleFile == ""},
		{"modelIndex", g.Globals.ModelIndexFile, "", g.Globals.ModelIndexFile == ""},
		{"functionIndex", g.Globals.FunctionIndexFile, "", g.Globals.FunctionIndexFile == ""},
		{"serviceIndex", g.Globals.ServiceIndexFile, "", !generateServices || g.Globals.ServiceIndexFile == ""},
		{"index", "index", "", !g.Options.IndexFile},
	}

	for _, f := range files {
		if f.skip {
			continue
		}
		var tmplData any = ctx
		if f.template == "modelIndex" {
			// modelIndex template needs modelIndex as a struct, not in map
			tmplData = ctx["modelIndex"]
		}
		if err := g.writeTemplate(f.template, tmplData, f.baseName, f.subDir); err != nil {
			return err
		}
	}

	// Sync temp to output
	if err := g.syncDirs(); err != nil {
		return fmt.Errorf("sync directories: %w", err)
	}

	g.Logger.Info("Generation finished with %d models and %d services.", len(g.Models), len(g.Services))
	return nil
}

func (g *Generator) mergeGlobals(ctx map[string]any) map[string]any {
	// Copy globals into context
	result := make(map[string]any)
	for k, v := range ctx {
		result[k] = v
	}
	result["configurationClass"] = g.Globals.ConfigurationClass
	result["configurationFile"] = g.Globals.ConfigurationFile
	result["configurationParams"] = g.Globals.ConfigurationParams
	result["baseServiceClass"] = g.Globals.BaseServiceClass
	result["baseServiceFile"] = g.Globals.BaseServiceFile
	result["apiServiceClass"] = g.Globals.ApiServiceClass
	result["apiServiceFile"] = g.Globals.ApiServiceFile
	result["requestBuilderClass"] = g.Globals.RequestBuilderClass
	result["requestBuilderFile"] = g.Globals.RequestBuilderFile
	result["responseClass"] = g.Globals.ResponseClass
	result["responseFile"] = g.Globals.ResponseFile
	result["moduleClass"] = g.Globals.ModuleClass
	result["moduleFile"] = g.Globals.ModuleFile
	result["modelIndexFile"] = g.Globals.ModelIndexFile
	result["functionIndexFile"] = g.Globals.FunctionIndexFile
	result["serviceIndexFile"] = g.Globals.ServiceIndexFile
	result["rootUrl"] = g.Globals.RootURL
	result["promises"] = g.Globals.Promises
	result["generateServices"] = g.Globals.GenerateServices
	result["module"] = g.Globals.ModuleClass
	return result
}

func (g *Generator) writeTemplate(name string, data any, baseName, subDir string) error {
	if g.TemplateM == nil {
		return fmt.Errorf("template manager not initialized")
	}
	// Convert data to map for template compatibility
	var tmplData map[string]any
	switch d := data.(type) {
	case map[string]any:
		tmplData = d
	case *model.Model:
		tmplData = modelToMap(d)
	case *operation.OperationVariant:
		tmplData = variantToMap(d)
	case *service.Service:
		tmplData = serviceToMap(d)
	default:
		tmplData = toMap(d)
	}
	// Merge globals so templates can access global config
	tmplData = g.mergeGlobals(tmplData)

	ts, err := g.TemplateM.Apply(name, tmplData)
	if err != nil {
		return fmt.Errorf("apply template %s: %w", name, err)
	}

	// Ensure the output ends with a newline (standard text file convention)
	if !strings.HasSuffix(ts, "\n") {
		ts += "\n"
	}

	dir := g.TempDir
	if subDir != "" {
		dir = filepath.Join(g.TempDir, subDir)
	}

	filePath := filepath.Join(dir, baseName+".ts")
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filePath, []byte(ts), 0644); err != nil {
		return fmt.Errorf("write %s: %w", filePath, err)
	}
	return nil
}

// sortedModels returns a sorted slice of models from the map
func sortedModels(models map[string]*model.Model) []*model.Model {
	names := make([]string, 0, len(models))
	for name := range models {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]*model.Model, len(names))
	for i, name := range names {
		result[i] = models[name]
	}
	return result
}

// ModelIndex is used for generating model index files
type ModelIndex struct {
	Imports    []*gen.Import
	PathToRoot string
}

func newModelIndex(models []*model.Model, opts *config.Options) *ModelIndex {
	mi := &ModelIndex{
		PathToRoot: "./",
	}
	imps := gen.NewImports(opts, "")
	for _, m := range models {
		imps.Add(m.Name, !m.IsEnum)
	}
	mi.Imports = imps.ToArray()
	for _, imp := range mi.Imports {
		imp.Path = "./models/"
		imp.FullPath = "models/" + imp.File
	}
	return mi
}

func (g *Generator) readRootURL() string {
	if g.Spec.Servers == nil || len(g.Spec.Servers) == 0 {
		return ""
	}
	server := g.Spec.Servers[0]
	rootURL := server.URL
	if rootURL == "" {
		return ""
	}
	for key, v := range server.Variables {
		rootURL = strings.ReplaceAll(rootURL, "{"+key+"}", v.Default)
	}
	return rootURL
}

func (g *Generator) readModels() {
	if g.Spec.Components == nil || g.Spec.Components.Schemas == nil {
		return
	}
	for name, sRaw := range g.Spec.Components.Schemas {
		sRaw := sRaw
		if sRaw.Ref != "" {
			schema, err := openapi.ResolveSchemaRef(g.Spec, &sRaw)
			if err != nil || schema == nil {
				continue
			}
			g.Models[name] = model.NewModel(g.Spec, name, schema, g.Options)
		} else {
			schema := sRaw.Schema
			g.Models[name] = model.NewModel(g.Spec, name, &schema, g.Options)
		}
	}
}

func (g *Generator) readServices() {
	defaultTag := g.Options.DefaultTag
	if defaultTag == "" {
		defaultTag = "Api"
	}

	opsByTag := make(map[string][]*operation.Operation)
	seenIDs := make(map[string]int)

	if g.Spec.Paths != nil {
		// Sort paths for deterministic output
		sortedPaths := make([]string, 0, len(g.Spec.Paths))
		for path := range g.Spec.Paths {
			sortedPaths = append(sortedPaths, path)
		}
		sort.Strings(sortedPaths)

		for _, path := range sortedPaths {
			pathSpec := g.Spec.Paths[path]
			if pathSpec == nil {
				continue
			}

			for _, method := range openapi.HTTPMethods {
				opSpec := getMethodOperation(pathSpec, method)
				if opSpec == nil {
					continue
				}

				id := opSpec.OperationID
				if id != "" {
					id = gen.MethodName(id)
				} else {
					id = gen.MethodName(path + "." + method)
					g.Logger.Warn("Operation '%s.%s' didn't specify an 'operationId'. Assuming '%s'.", path, method, id)
				}

				// Handle duplicates
				if _, exists := seenIDs[id]; exists {
					seenIDs[id]++
					newID := fmt.Sprintf("%s_%d", id, seenIDs[id])
					g.Logger.Warn("Duplicate operation id '%s'. Assuming id %s for operation '%s.%s'.", id, newID, path, method)
					id = newID
				} else {
					seenIDs[id] = 0
				}

				op := operation.NewOperation(g.Spec, path, pathSpec, method, id, opSpec, g.Options)

				// Default tag
				if len(op.Tags) == 0 {
					g.Logger.Warn("No tags set on operation '%s.%s'. Assuming '%s'.", path, method, defaultTag)
					op.Tags = append(op.Tags, defaultTag)
				}

				for _, tag := range op.Tags {
					opsByTag[tag] = append(opsByTag[tag], op)
				}
			}
		}

		// Filter tags
		includeTags := g.Options.IncludeTags
		excludeTags := g.Options.ExcludeTags

		// Sort tags for deterministic output
		sortedTags := make([]string, 0, len(opsByTag))
		for tag := range opsByTag {
			sortedTags = append(sortedTags, tag)
		}
		sort.Strings(sortedTags)

		for _, tagName := range sortedTags {
			ops := opsByTag[tagName]
			if len(includeTags) > 0 && !contains(includeTags, tagName) {
				g.Logger.Info("Ignoring tag %s because it is not listed in the 'includeTags' option", tagName)
				continue
			}
			if len(excludeTags) > 0 && contains(excludeTags, tagName) {
				g.Logger.Info("Ignoring tag %s because it is listed in the 'excludeTags' option", tagName)
				continue
			}

			// Get tag description
			tagDesc := ""
			if g.Spec.Tags != nil {
				for _, t := range g.Spec.Tags {
					if t.Name == tagName {
						tagDesc = t.Description
						break
					}
				}
			}

			svc := service.NewService(tagName, tagDesc, ops, g.Options)
			g.Services[tagName] = svc
		}
	}
}

func getMethodOperation(pi *openapi.PathItem, method string) *openapi.Operation {
	switch method {
	case "get":
		return pi.Get
	case "put":
		return pi.Put
	case "post":
		return pi.Post
	case "delete":
		return pi.Delete
	case "options":
		return pi.Options
	case "head":
		return pi.Head
	case "patch":
		return pi.Patch
	case "trace":
		return pi.Trace
	}
	return nil
}

func (g *Generator) validateVersion() {
	version := g.Spec.OpenAPI
	if version == "" {
		g.Logger.Warn("OpenAPI specification version is missing")
		return
	}
	if strings.HasPrefix(version, "3.0") || strings.HasPrefix(version, "3.1") {
		g.Logger.Info("Using OpenAPI specification version: %s", version)
	} else {
		g.Logger.Warn("Unsupported OpenAPI version: %s. Only 3.0.x and 3.1.x are supported.", version)
	}
}

func (g *Generator) ignoreUnusedModels() {
	usedNames := make(map[string]bool)

	for _, svc := range g.Services {
		for _, imp := range svc.Imports {
			if strings.Contains(imp.Path, "models/") {
				usedNames[imp.Name] = true
			}
		}
		for _, op := range svc.Operations {
			for _, variant := range op.Variants {
				for _, imp := range variant.Imports {
					if strings.Contains(imp.Path, "models/") {
						usedNames[imp.Name] = true
					}
				}
			}
		}
	}

	// Collect transitive dependencies
	referenced := make([]string, 0, len(usedNames))
	for name := range usedNames {
		referenced = append(referenced, name)
	}
	usedNames = make(map[string]bool)
	for _, name := range referenced {
		g.collectDeps(name, usedNames)
	}

	for name := range g.Models {
		if !usedNames[name] {
			g.Logger.Debug("Ignoring model %s because it is not used anywhere", name)
			delete(g.Models, name)
		}
	}
}

func (g *Generator) collectDeps(name string, used map[string]bool) {
	m, exists := g.Models[name]
	if !exists || used[name] {
		return
	}
	used[name] = true
	for _, refName := range g.allRefNames(m.Schema) {
		g.collectDeps(refName, used)
	}
}

func (g *Generator) allRefNames(schema *openapi.Schema) []string {
	if schema == nil {
		return nil
	}
	var result []string
	for _, s := range schema.AllOf {
		s := s
		if s.Ref != "" {
			result = append(result, gen.SimpleName(s.Ref))
		} else {
			result = append(result, g.allRefNames(&s.Schema)...)
		}
	}
	for _, s := range schema.OneOf {
		s := s
		if s.Ref != "" {
			result = append(result, gen.SimpleName(s.Ref))
		} else {
			result = append(result, g.allRefNames(&s.Schema)...)
		}
	}
	for _, s := range schema.AnyOf {
		s := s
		if s.Ref != "" {
			result = append(result, gen.SimpleName(s.Ref))
		} else {
			result = append(result, g.allRefNames(&s.Schema)...)
		}
	}
	for name, prop := range schema.Properties {
		_ = name
		prop := prop
		if prop.Ref != "" {
			result = append(result, gen.SimpleName(prop.Ref))
		} else {
			result = append(result, g.allRefNames(&prop.Schema)...)
		}
	}
	if schema.Items != nil {
		if schema.Items.Ref != "" {
			result = append(result, gen.SimpleName(schema.Items.Ref))
		} else {
			result = append(result, g.allRefNames(&schema.Items.Schema)...)
		}
	}
	return result
}

func (g *Generator) syncDirs() error {
	if err := os.MkdirAll(g.OutDir, 0755); err != nil {
		return err
	}

	removeStale := g.Options.RemoveStaleFiles == nil || *g.Options.RemoveStaleFiles

	// Walk temp dir and copy files
	if err := filepath.Walk(g.TempDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(g.TempDir, srcPath)
		if err != nil {
			return err
		}
		destPath := filepath.Join(g.OutDir, rel)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		data, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}

		existing, err := os.ReadFile(destPath)
		if err != nil || string(data) != string(existing) {
			if err := os.WriteFile(destPath, data, 0644); err != nil {
				return err
			}
			g.Logger.Debug("Wrote %s", destPath)
		}
		return nil
	}); err != nil {
		return err
	}

	// Remove stale files
	if removeStale {
		if err := filepath.Walk(g.OutDir, func(destPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(g.OutDir, destPath)
			if err != nil {
				return err
			}
			srcPath := filepath.Join(g.TempDir, rel)
			if _, err := os.Stat(srcPath); os.IsNotExist(err) {
				if err := os.Remove(destPath); err != nil {
					return err
				}
				g.Logger.Debug("Removed stale file %s", destPath)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func getOrDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
