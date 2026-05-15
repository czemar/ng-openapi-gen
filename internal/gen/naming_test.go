package gen

import (
	"testing"

	"github.com/czemar/ng-openapi-gen/internal/config"
)

func TestSimpleName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"#/components/schemas/Pet", "Pet"},
		{"Pet", "Pet"},
		{"", ""},
		{"a/b/c", "c"},
	}
	for _, tt := range tests {
		got := SimpleName(tt.input)
		if got != tt.expected {
			t.Errorf("SimpleName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNamespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"com.example.Pet", "com/example"},
		{"Pet", ""},
		{"", ""},
		{".a.b.C", "a/b"},
	}
	for _, tt := range tests {
		got := Namespace(tt.input)
		if got != tt.expected {
			t.Errorf("Namespace(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestUpperFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"Hello", "Hello"},
		{"", ""},
		{"a", "A"},
		{"123", "123"},
	}
	for _, tt := range tests {
		got := UpperFirst(tt.input)
		if got != tt.expected {
			t.Errorf("UpperFirst(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", "helloWorld"},
		{"HelloWorld", "helloWorld"},
		{"hello-world", "helloWorld"},
		{"HELLO", "hello"},
		{"", ""},
		{"single", "single"},
		{"alreadyCamel", "alreadyCamel"},
		{"snake_case_name", "snakeCaseName"},
	}
	for _, tt := range tests {
		got := CamelCase(tt.input)
		if got != tt.expected {
			t.Errorf("CamelCase(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"helloWorld", "hello-world"},
		{"HelloWorld", "hello-world"},
		{"hello_world", "hello-world"},
		{"HELLO", "hello"},
		{"", ""},
		{"already-kebab", "already-kebab"},
	}
	for _, tt := range tests {
		got := KebabCase(tt.input)
		if got != tt.expected {
			t.Errorf("KebabCase(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PetModel", "pet-model"},
		{"listPets", "list-pets"},
		{"ApiConfiguration", "api-configuration"},
		{"StrictHttpResponse", "strict-http-response"},
		{"RequestBuilder", "request-builder"},
		{"BaseService", "base-service"},
	}
	for _, tt := range tests {
		got := FileName(tt.input)
		if got != tt.expected {
			t.Errorf("FileName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMethodName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"list-pets", "listPets"},
		{"get_pet_by_id", "getPetById"},
		{"CreatePets", "createPets"},
		{"FOO", "foo"},
		{"", ""},
	}
	for _, tt := range tests {
		got := MethodName(tt.input)
		if got != tt.expected {
			t.Errorf("MethodName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestTypeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", "HelloWorld"},
		{"list-pets", "ListPets"},
		{"pet", "Pet"},
		{"", ""},
	}
	for _, tt := range tests {
		got := TypeName(tt.input, nil)
		if got != tt.expected {
			t.Errorf("TypeName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestModelClass(t *testing.T) {
	opts := &config.Options{
		ModelPrefix: "Petstore",
		ModelSuffix: "Model",
	}
	got := ModelClass("Pet", opts)
	expected := "PetstorePetModel"
	if got != expected {
		t.Errorf("ModelClass = %q, want %q", got, expected)
	}

	opts2 := &config.Options{}
	got2 := ModelClass("Pet", opts2)
	if got2 != "Pet" {
		t.Errorf("ModelClass without prefix/suffix = %q, want %q", got2, "Pet")
	}
}

func TestServiceClass(t *testing.T) {
	opts := &config.Options{
		ServicePrefix: "My",
	}
	got := ServiceClass("Pet", opts)
	expected := "MyPetService"
	if got != expected {
		t.Errorf("ServiceClass = %q, want %q", got, expected)
	}

	opts2 := &config.Options{}
	got2 := ServiceClass("Pet", opts2)
	if got2 != "PetService" {
		t.Errorf("ServiceClass default suffix = %q, want %q", got2, "PetService")
	}
}

func TestEnsureNotReserved(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"class", "class$"},
		{"return", "return$"},
		{"myVar", "myVar"},
		{"", ""},
	}
	for _, tt := range tests {
		got := EnsureNotReserved(tt.input)
		if got != tt.expected {
			t.Errorf("EnsureNotReserved(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestEscapeId(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"validId", "validId"},
		{"123abc", "'123abc'"},
		{"has spaces", "'has spaces'"},
		{"it's", "'it\\'s'"},
	}
	for _, tt := range tests {
		got := EscapeId(tt.input)
		if got != tt.expected {
			t.Errorf("EscapeId(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestToBasicChars(t *testing.T) {
	tests := []struct {
		input         string
		firstNonDigit bool
		expected      string
	}{
		{"hello world", true, "hello_world"},
		{"123abc", true, "_123abc"},
		{"123abc", false, "123abc"},
		{"foo-bar", true, "foo_bar"},
		{"café", true, "caf_"},
	}
	for _, tt := range tests {
		got := ToBasicChars(tt.input, tt.firstNonDigit)
		if got != tt.expected {
			t.Errorf("ToBasicChars(%q, %v) = %q, want %q", tt.input, tt.firstNonDigit, got, tt.expected)
		}
	}
}

func TestEnumName(t *testing.T) {
	tests := []struct {
		value    string
		style    string
		expected string
	}{
		{"vanilla", "alias", "Vanilla"},
		{"StrawBerry", "alias", "StrawBerry"},
		{"cookie dough", "alias", "CookieDough"},
		{"vanilla", "upper", "VANILLA"},
		{"StrawBerry", "upper", "STRAWBERRY"},
		{"cookie dough", "upper", "COOKIE_DOUGH"},
		{"butter_pecan", "upper", "BUTTER_PECAN"},
		{"vanilla", "ignorecase", "vanilla"},
		{"StrawBerry", "ignorecase", "StrawBerry"},
		{"cookie dough", "ignorecase", "cookie_dough"},
		{"123abc", "alias", "$123abc"},
	}
	for _, tt := range tests {
		opts := &config.Options{EnumStyle: tt.style}
		got := EnumName(tt.value, opts)
		if got != tt.expected {
			t.Errorf("EnumName(%q, %q) = %q, want %q", tt.value, tt.style, got, tt.expected)
		}
	}
}

func TestPathToRoot(t *testing.T) {
	tests := []struct {
		depth    int
		expected string
	}{
		{0, "./"},
		{1, "../"},
		{2, "../../"},
		{3, "../../../"},
	}
	for _, tt := range tests {
		got := PathToRoot(tt.depth)
		if got != tt.expected {
			t.Errorf("PathToRoot(%d) = %q, want %q", tt.depth, got, tt.expected)
		}
	}
}
