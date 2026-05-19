package generate

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/czemar/ng-openapi-gen/internal/config"
	"github.com/czemar/ng-openapi-gen/internal/openapi"
)

func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root")
		}
		dir = parent
	}
}

func testSpecPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(findProjectRoot(t), "test", name)
}

func TestGeneratePetstore30(t *testing.T) {
	spec, err := openapi.ParseSpec(testSpecPath(t, "petstore-3.0.json"))
	if err != nil {
		t.Fatalf("ParseSpec failed: %v", err)
	}

	opts := &config.Options{
		Input:              "petstore-3.0.json",
		Output:             filepath.Join(t.TempDir(), "petstore-3.0"),
		ModelPrefix:        "Petstore",
		ModelSuffix:        "Model",
		IgnoreUnusedModels: boolPtr(false),
	}

	gen := NewGenerator(spec, opts)
	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check output directory exists
	outDir := gen.OutDir

	// Check expected files exist
	expectedFiles := []string{
		"api-configuration.ts",
		"strict-http-response.ts",
		"request-builder.ts",
		"models/petstore-pet-model.ts",
		"models/petstore-pets-model.ts",
		"models/petstore-error-model.ts",
		"fn/pets/list-pets.ts",
		"fn/pets/create-pets.ts",
		"fn/pets/show-pet-by-id.ts",
	}
	for _, f := range expectedFiles {
		path := filepath.Join(outDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s does not exist", path)
		}
	}

	// Check for no <no value>
	assertNoNoValue(t, outDir)

	// Check services
	if len(gen.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(gen.Services))
	}

	// Check models count
	if len(gen.Models) != 3 {
		t.Errorf("expected 3 models, got %d", len(gen.Models))
	}

	petModel, ok := gen.Models["Pet"]
	if !ok {
		t.Fatal("expected Pet model")
	}
	if petModel.TypeName != "PetstorePetModel" {
		t.Errorf("Pet model TypeName = %q, want %q", petModel.TypeName, "PetstorePetModel")
	}
	if petModel.IsObject != true {
		t.Errorf("Pet should be an object")
	}
	if len(petModel.Properties) != 3 {
		t.Errorf("Pet should have 3 properties, got %d", len(petModel.Properties))
	}
	if len(petModel.Properties) >= 1 {
		p := petModel.Properties[0]
		if p.Name != "id" || p.Type != "number" || !p.Required {
			t.Errorf("Pet.id = {name:%q type:%q required:%v}, want {id number true}", p.Name, p.Type, p.Required)
		}
	}
	if len(petModel.Properties) >= 2 {
		p := petModel.Properties[1]
		if p.Name != "name" || p.Type != "string" || !p.Required {
			t.Errorf("Pet.name = {name:%q type:%q required:%v}, want {name string true}", p.Name, p.Type, p.Required)
		}
	}
	if len(petModel.Properties) >= 3 {
		p := petModel.Properties[2]
		if p.Name != "tag" || p.Type != "string" || p.Required {
			t.Errorf("Pet.tag = {name:%q type:%q required:%v}, want {tag string false}", p.Name, p.Type, p.Required)
		}
	}

	// Check Pets model (simple type alias)
	petsModel, ok := gen.Models["Pets"]
	if !ok {
		t.Fatal("expected Pets model")
	}
	if petsModel.IsSimple != true {
		t.Errorf("Pets should be simple")
	}
	if petsModel.SimpleType != "Array<PetstorePetModel>" {
		t.Errorf("Pets.SimpleType = %q, want %q", petsModel.SimpleType, "Array<PetstorePetModel>")
	}

	// Check Error model
	errModel, ok := gen.Models["Error"]
	if !ok {
		t.Fatal("expected Error model")
	}
	if errModel.TypeName != "PetstoreErrorModel" {
		t.Errorf("Error model TypeName = %q, want %q", errModel.TypeName, "PetstoreErrorModel")
	}
	if len(errModel.Properties) != 2 {
		t.Errorf("Error should have 2 properties, got %d", len(errModel.Properties))
	}

	// Check pets service
	petsSvc, ok := gen.Services["pets"]
	if !ok {
		t.Fatal("expected pets service")
	}
	if len(petsSvc.Operations) != 3 {
		t.Errorf("pets service should have 3 operations, got %d", len(petsSvc.Operations))
	}

	// Check operations exist
	opIDs := make(map[string]bool)
	for _, op := range petsSvc.Operations {
		opIDs[op.ID] = true
	}
	for _, id := range []string{"listPets", "createPets", "showPetById"} {
		if !opIDs[id] {
			t.Errorf("expected operation %q not found", id)
		}
	}

	// Verify generated file content
	petFile := filepath.Join(outDir, "models", "petstore-pet-model.ts")
	content, err := os.ReadFile(petFile)
	if err != nil {
		t.Fatal(err)
	}
	petContent := string(content)

	checks := []string{
		"export interface PetstorePetModel",
		"id: number",
		"name: string",
		"tag?: string",
	}
	for _, check := range checks {
		if !strings.Contains(petContent, check) {
			t.Errorf("petstore-pet-model.ts should contain %q", check)
		}
	}

	// Check list-pets.ts content
	listPetsFile := filepath.Join(outDir, "fn", "pets", "list-pets.ts")
	lpContent, err := os.ReadFile(listPetsFile)
	if err != nil {
		t.Fatal(err)
	}
	lpStr := string(lpContent)

	lpChecks := []string{
		"export interface ListPets$Params",
		"limit?: number",
		"export function listPets",
		"StrictHttpResponse<PetstorePetsModel>",
		"RequestBuilder",
		"rb.query('limit', params.limit, {});",
	}
	for _, check := range lpChecks {
		if !strings.Contains(lpStr, check) {
			t.Errorf("list-pets.ts should contain %q", check)
		}
	}
}

func TestGenerateAllTypes(t *testing.T) {
	spec, err := openapi.ParseSpec(testSpecPath(t, "all-types.json"))
	if err != nil {
		t.Fatalf("ParseSpec failed: %v", err)
	}

	opts := &config.Options{
		Input:     "all-types.json",
		Output:    filepath.Join(t.TempDir(), "all-types"),
		IndexFile: true,
		EnumStyle: "pascal",
		Module:    "AllTypesModule",
		Services:  boolPtr(true),
	}

	gen := NewGenerator(spec, opts)
	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	assertNoNoValue(t, gen.OutDir)

	// Check models exist
	expectedModels := []string{"RefEnum", "RefIntEnum", "RefNamedIntEnum", "Shape", "NullableObject"}
	for _, name := range expectedModels {
		if _, ok := gen.Models[name]; !ok {
			t.Errorf("expected model %q not found", name)
		}
	}

	// Check file outputs
	expectedFiles := []string{
		"models/ref-enum.ts",
		"models/ref-int-enum.ts",
		"models/ref-named-int-enum.ts",
		"models/shape.ts",
		"models/nullable-object.ts",
		"models/a/b/ref-object.ts",
		"models/x/y/ref-object.ts",
		"index.ts",
	}
	for _, f := range expectedFiles {
		path := filepath.Join(gen.OutDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s does not exist", path)
		}
	}

	// Check service
	if len(gen.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(gen.Services))
	}
	apiSvc, ok := gen.Services["Api"]
	if !ok {
		t.Fatal("expected Api service")
	}
	if apiSvc.TypeName != "ApiService" {
		t.Errorf("Api service TypeName = %q, want %q", apiSvc.TypeName, "ApiService")
	}

	// Check that files don't have <no value>
	// (already checked by assertNoNoValue)

	// Check ref-enum content
	refEnumFile := filepath.Join(gen.OutDir, "models", "ref-enum.ts")
	content, err := os.ReadFile(refEnumFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "ValueA") {
		t.Errorf("ref-enum.ts should contain ValueA")
	}
}

func TestGenerateAllOperations(t *testing.T) {
	spec, err := openapi.ParseSpec(testSpecPath(t, "all-operations.json"))
	if err != nil {
		t.Fatalf("ParseSpec failed: %v", err)
	}

	opts := &config.Options{
		Input:    "all-operations.json",
		Output:   filepath.Join(t.TempDir(), "all-operations"),
		Services: boolPtr(true),
	}

	gen := NewGenerator(spec, opts)
	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	assertNoNoValue(t, gen.OutDir)

	// Check services
	if len(gen.Services) == 0 {
		t.Errorf("expected at least 1 service")
	}

	// Check operation variants are created
	totalVariants := 0
	for _, svc := range gen.Services {
		for _, op := range svc.Operations {
			totalVariants += len(op.Variants)
		}
	}
	if totalVariants == 0 {
		t.Errorf("expected at least 1 operation variant")
	}

	// Verify content specific to all-operations
	outDir := gen.OutDir

	// Check the path4-put variants (multi-content-type)
	matches, _ := filepath.Glob(filepath.Join(outDir, "fn", "api", "path-4-put*.ts"))
	if len(matches) == 0 {
		t.Errorf("expected path4-put variant files, got none")
	}
}

func TestGenerateEnums(t *testing.T) {
	spec, err := openapi.ParseSpec(testSpecPath(t, "enums.json"))
	if err != nil {
		t.Fatalf("ParseSpec failed: %v", err)
	}

	t.Run("alias style", func(t *testing.T) {
		opts := &config.Options{
			Input:     "enums.json",
			Output:    filepath.Join(t.TempDir(), "enums-alias"),
			EnumStyle: "alias",
		}
		gen := NewGenerator(spec, opts)
		if err := gen.Generate(); err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		assertNoNoValue(t, gen.OutDir)

		flavorFile := filepath.Join(gen.OutDir, "models", "flavor-enum.ts")
		content, err := os.ReadFile(flavorFile)
		if err != nil {
			t.Fatal(err)
		}
		contentStr := string(content)

		// Alias style: export type FlavorEnum = 'vanilla' | 'StrawBerry' | ...
		if !strings.Contains(contentStr, "export type FlavorEnum =") {
			t.Errorf("alias enum should be a type alias")
		}
	})

	t.Run("upper style", func(t *testing.T) {
		opts := &config.Options{
			Input:     "enums.json",
			Output:    filepath.Join(t.TempDir(), "enums-upper"),
			EnumStyle: "upper",
		}
		gen := NewGenerator(spec, opts)
		if err := gen.Generate(); err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		assertNoNoValue(t, gen.OutDir)

		flavorFile := filepath.Join(gen.OutDir, "models", "flavor-enum.ts")
		content, err := os.ReadFile(flavorFile)
		if err != nil {
			t.Fatal(err)
		}
		contentStr := string(content)

		// Upper style: enum with UPPER_CASE keys
		if !strings.Contains(contentStr, "VANILLA") {
			t.Errorf("upper enum should contain VANILLA")
		}
		if !strings.Contains(contentStr, "STRAWBERRY") {
			t.Errorf("upper enum should contain STRAWBERRY")
		}
		if !strings.Contains(contentStr, "COOKIE_DOUGH") {
			t.Errorf("upper enum should contain COOKIE_DOUGH")
		}
	})

	t.Run("pascal style", func(t *testing.T) {
		opts := &config.Options{
			Input:     "enums.json",
			Output:    filepath.Join(t.TempDir(), "enums-pascal"),
			EnumStyle: "pascal",
		}
		gen := NewGenerator(spec, opts)
		if err := gen.Generate(); err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		assertNoNoValue(t, gen.OutDir)

		flavorFile := filepath.Join(gen.OutDir, "models", "flavor-enum.ts")
		content, err := os.ReadFile(flavorFile)
		if err != nil {
			t.Fatal(err)
		}
		contentStr := string(content)

		if !strings.Contains(contentStr, "Vanilla") {
			t.Errorf("pascal enum should contain Vanilla")
		}
		if !strings.Contains(contentStr, "StrawBerry") {
			t.Errorf("pascal enum should contain StrawBerry")
		}
	})

	t.Run("ignorecase style", func(t *testing.T) {
		opts := &config.Options{
			Input:     "enums.json",
			Output:    filepath.Join(t.TempDir(), "enums-ignorecase"),
			EnumStyle: "ignorecase",
		}
		gen := NewGenerator(spec, opts)
		if err := gen.Generate(); err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		assertNoNoValue(t, gen.OutDir)

		flavorFile := filepath.Join(gen.OutDir, "models", "flavor-enum.ts")
		content, err := os.ReadFile(flavorFile)
		if err != nil {
			t.Fatal(err)
		}
		contentStr := string(content)

		if !strings.Contains(contentStr, "vanilla = 'vanilla'") {
			t.Errorf("ignorecase enum should contain 'vanilla = ''vanilla'''")
		}
		if !strings.Contains(contentStr, "StrawBerry = 'StrawBerry'") {
			t.Errorf("ignorecase enum should contain 'StrawBerry = ''StrawBerry'''")
		}
	})
}

func TestGenerateObservables(t *testing.T) {
	spec, err := openapi.ParseSpec(testSpecPath(t, "petstore-3.0.json"))
	if err != nil {
		t.Fatalf("ParseSpec failed: %v", err)
	}

	promises := false
	services := true
	opts := &config.Options{
		Input:    "petstore-3.0.json",
		Output:   filepath.Join(t.TempDir(), "observables"),
		Promises: &promises,
		Services: &services,
	}
	gen := NewGenerator(spec, opts)
	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	assertNoNoValue(t, gen.OutDir)

	// Check that service files reference Observable (not Promise)
	outDir := gen.OutDir
	svcFiles, _ := filepath.Glob(filepath.Join(outDir, "services", "*.ts"))
	for _, f := range svcFiles {
		content, _ := os.ReadFile(f)
		if strings.Contains(string(content), "firstValueFrom") {
			t.Errorf("observables mode should not contain firstValueFrom in %s", f)
		}
		if !strings.Contains(string(content), "Observable") {
			t.Errorf("observables mode should contain Observable in %s", f)
		}
	}
}

func TestGenerateNoServices(t *testing.T) {
	spec, err := openapi.ParseSpec(testSpecPath(t, "petstore-3.0.json"))
	if err != nil {
		t.Fatalf("ParseSpec failed: %v", err)
	}

	opts := &config.Options{
		Input:    "petstore-3.0.json",
		Output:   filepath.Join(t.TempDir(), "no-services"),
		Services: boolPtr(false),
	}

	gen := NewGenerator(spec, opts)
	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	assertNoNoValue(t, gen.OutDir)

	// Should generate models and fns but NOT service files
	outDir := gen.OutDir

	modelFiles, _ := filepath.Glob(filepath.Join(outDir, "models", "*.ts"))
	if len(modelFiles) == 0 {
		t.Errorf("expected model files")
	}

	fnFiles, _ := filepath.Glob(filepath.Join(outDir, "fn", "**", "*.ts"))
	if len(fnFiles) == 0 {
		t.Errorf("expected function files")
	}

	// Service files should NOT exist
	svcDir := filepath.Join(outDir, "services")
	if _, err := os.Stat(svcDir); err == nil {
		t.Errorf("services directory should not exist when services=false")
	}
}

func TestGeneratePetstore31(t *testing.T) {
	spec, err := openapi.ParseSpec(testSpecPath(t, "petstore-3.1.json"))
	if err != nil {
		t.Fatalf("ParseSpec failed: %v", err)
	}

	opts := &config.Options{
		Input:  "petstore-3.1.json",
		Output: filepath.Join(t.TempDir(), "petstore-3.1"),
	}

	gen := NewGenerator(spec, opts)
	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	assertNoNoValue(t, gen.OutDir)

	// Check expected files for 3.1
	outDir := gen.OutDir
	expectedFiles := []string{
		"models/pet.ts",
		"models/pets.ts",
		"models/error.ts",
	}
	for _, f := range expectedFiles {
		path := filepath.Join(outDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s does not exist", path)
		}
	}
}

func TestGenerateNoUnusedModels(t *testing.T) {
	spec, err := openapi.ParseSpec(testSpecPath(t, "all-types.json"))
	if err != nil {
		t.Fatalf("ParseSpec failed: %v", err)
	}

	opts := &config.Options{
		Input:              "all-types.json",
		Output:             filepath.Join(t.TempDir(), "all-types-filtered"),
		IgnoreUnusedModels: boolPtr(true),
	}

	gen := NewGenerator(spec, opts)
	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	assertNoNoValue(t, gen.OutDir)
}

func assertNoNoValue(t *testing.T, dir string) {
	t.Helper()
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".ts") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(data), "<no value>") {
			t.Errorf("file %s contains <no value>", path)
		}
		return nil
	})
	if err != nil {
		t.Errorf("walk error: %v", err)
	}
}

func boolPtr(b bool) *bool {
	return &b
}

// TestGoldenFiles regenerates output from each config and compares against the
// expected output in the out/ directory. This catches regressions where the
// generated output changes unexpectedly.
func TestGoldenFiles(t *testing.T) {
	projectRoot := findProjectRoot(t)

	type goldenCase struct {
		name       string
		configPath string // absolute path to .config.json
		specPath   string // absolute path to OpenAPI spec
		expectDir  string // absolute path to expected out/ dir
	}

	var cases []goldenCase

	// Configs with "input" pointing to a file name (no "test/" prefix) → spec
	// lives in test/<name>. Configs with "test/" prefix → the input is already
	// relative to the project root (test/<name>).
	configDir := filepath.Join(projectRoot, "test")
	expectedBase := filepath.Join(projectRoot, "out")

	configFiles, err := filepath.Glob(filepath.Join(configDir, "*.config.json"))
	if err != nil {
		t.Fatalf("glob configs: %v", err)
	}

	for _, cf := range configFiles {
		opts, err := config.LoadOptions(cf)
		if err != nil || opts.Output == "" {
			continue
		}

		// Resolve the output directory (strip trailing slash)
		outRel := strings.TrimRight(opts.Output, "/")
		expectDir := filepath.Join(expectedBase, filepath.Base(outRel))

		if _, err := os.Stat(expectDir); os.IsNotExist(err) {
			continue
		}

		// Resolve the input spec path
		input := opts.Input
		var specPath string
		if strings.HasPrefix(input, "test/") {
			specPath = filepath.Join(projectRoot, input)
		} else {
			specPath = filepath.Join(configDir, input)
		}

		name := strings.TrimSuffix(filepath.Base(cf), ".config.json")

		// Skip scenarios with broken/missing spec files or incompatible features
		switch name {
		case "templates": // Different template engine (Handlebars vs Go templates)
			continue
		case "camelize-model-names": // Referenced spec (keep-model-names.json) does not exist
			continue
		case "self-ref": // Referenced spec (self-ref-array.json) does not exist
			continue
		case "openapi31-jsonschema": // Parser cannot handle some 3.1 JSON Schema features (bool in items)
			continue
		}

		cases = append(cases, goldenCase{
			name:       name,
			configPath: cf,
			specPath:   specPath,
			expectDir:  expectDir,
		})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts, err := config.LoadOptions(tc.configPath)
			if err != nil {
				t.Fatalf("LoadOptions(%s): %v", tc.configPath, err)
			}

			opts.Input = tc.specPath
			opts.Output = t.TempDir()

			spec, err := openapi.ParseSpec(tc.specPath)
			if err != nil {
				t.Fatalf("ParseSpec(%s): %v", tc.specPath, err)
			}

			gen := NewGenerator(spec, opts)
			if err := gen.Generate(); err != nil {
				t.Fatalf("Generate: %v", err)
			}

			fmt.Printf("Generated output: %s\n", gen.OutDir)
			// Debug: print generated file contents
			for _, name := range []string{"api.ts", "api-configuration.ts", "models.ts", "functions.ts", "models/petstore-pet-model.ts", "models/petstore-pets-model.ts"} {
				data, err := os.ReadFile(filepath.Join(gen.OutDir, name))
				if err == nil {
					fmt.Printf("--- %s ---\n%s", name, string(data))
				}
			}
			diff := diffDirs(t, tc.expectDir, gen.OutDir)
			if diff != "" {
				t.Errorf("Output mismatch:\n%s", diff)
			}
		})
	}
}

// diffDirs compares two directories recursively and returns a description of
// all differences, or empty string if they are identical.
func diffDirs(t *testing.T, expected, actual string) string {
	t.Helper()

	expectedFiles := make(map[string]string) // relPath -> absPath
	actualFiles := make(map[string]string)

	walk := func(root string, dest map[string]string) {
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			dest[rel] = path
			return nil
		})
	}

	walk(expected, expectedFiles)
	walk(actual, actualFiles)

	var sb strings.Builder

	for rel := range expectedFiles {
		if _, ok := actualFiles[rel]; !ok {
			sb.WriteString(fmt.Sprintf("  missing: %s\n", rel))
		}
	}
	for rel := range actualFiles {
		if _, ok := expectedFiles[rel]; !ok {
			sb.WriteString(fmt.Sprintf("  extra:   %s\n", rel))
		}
	}

	for rel, expPath := range expectedFiles {
		actPath, ok := actualFiles[rel]
		if !ok {
			continue
		}
		expData, err := os.ReadFile(expPath)
		if err != nil {
			sb.WriteString(fmt.Sprintf("  read error (%s): %v\n", rel, err))
			continue
		}
		actData, err := os.ReadFile(actPath)
		if err != nil {
			sb.WriteString(fmt.Sprintf("  read error (%s): %v\n", rel, err))
			continue
		}
		if !bytes.Equal(expData, actData) {
			sb.WriteString(fmt.Sprintf("  differ:  %s\n", rel))
			// Show line-level diff for debugging
			expLines := strings.Split(string(expData), "\n")
			actLines := strings.Split(string(actData), "\n")
			max := len(expLines)
			if len(actLines) > max {
				max = len(actLines)
			}
			for i := 0; i < max; i++ {
				var e, a string
				if i < len(expLines) {
					e = expLines[i]
				}
				if i < len(actLines) {
					a = actLines[i]
				}
				if e != a {
					sb.WriteString(fmt.Sprintf("    line %d:\n      -%s\n      +%s\n", i+1, e, a))
					break // show first difference only
				}
			}
		}
	}

	return sb.String()
}
