package gen

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/czemar/ng-openapi-gen/internal/config"
)

var reservedKeywords = map[string]bool{
	"abstract": true, "arguments": true, "await": true, "boolean": true,
	"break": true, "byte": true, "case": true, "catch": true, "char": true,
	"class": true, "const": true, "continue": true, "debugger": true,
	"default": true, "delete": true, "do": true, "double": true, "else": true,
	"enum": true, "eval": true, "export": true, "extends": true, "false": true,
	"final": true, "finally": true, "float": true, "for": true, "function": true,
	"goto": true, "if": true, "implements": true, "import": true, "in": true,
	"instanceof": true, "int": true, "interface": true, "let": true, "long": true,
	"native": true, "new": true, "null": true, "package": true, "private": true,
	"protected": true, "public": true, "return": true, "short": true,
	"static": true, "super": true, "switch": true, "synchronized": true,
	"this": true, "throw": true, "throws": true, "transient": true, "true": true,
	"try": true, "typeof": true, "var": true, "void": true, "volatile": true,
	"while": true, "with": true, "yield": true,
}

// SimpleName returns the last part after '/'
func SimpleName(name string) string {
	pos := strings.LastIndex(name, "/")
	if pos >= 0 {
		return name[pos+1:]
	}
	return name
}

// Namespace returns the part before the last '.', split by '/' instead of '.'
func Namespace(name string) string {
	name = strings.TrimLeft(name, ".")
	name = strings.TrimRight(name, ".")
	pos := strings.LastIndex(name, ".")
	if pos < 0 {
		return ""
	}
	return strings.ReplaceAll(name[:pos], ".", "/")
}

// UnqualifiedName returns the last part after '.' with model prefix/suffix
func UnqualifiedName(name string, opts *config.Options) string {
	pos := strings.LastIndex(name, ".")
	if pos >= 0 {
		return ModelClass(name[pos+1:], opts)
	}
	return ModelClass(name, opts)
}

// QualifiedName returns the namespace-qualified type name
func QualifiedName(name string, opts *config.Options) string {
	ns := Namespace(name)
	unq := UnqualifiedName(name, opts)
	if ns != "" {
		return TypeName(ns, opts) + unq
	}
	return unq
}

// TypeName returns a suitable TypeScript type/class name
func TypeName(name string, opts *config.Options) string {
	if opts != nil && opts.CamelizeModelNames != nil && !*opts.CamelizeModelNames {
		return UpperFirst(ToBasicChars(name, true))
	}
	return UpperFirst(MethodName(name))
}

// MethodName returns a camelCase method name
func MethodName(name string) string {
	return CamelCase(ToBasicChars(name, true))
}

// FileName returns a kebab-case file name
func FileName(text string) string {
	return KebabCase(ToBasicChars(text, false))
}

// EnumName returns the enum constant name for a given value
func EnumName(value string, opts *config.Options) string {
	name := ToBasicChars(value, true)
	switch opts.EnumStyle {
	case "ignorecase":
		// keep as-is
	case "upper":
		name = strings.ToUpper(regexp.MustCompile(`\s+`).ReplaceAllString(name, "_"))
		// Convert camelCase to UPPER_CASE with underscores
		var result strings.Builder
		for i, r := range name {
			if i > 0 && unicode.IsUpper(r) && (i+1 < len(name) && unicode.IsLower(rune(name[i+1])) || unicode.IsLower(rune(name[i-1]))) {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToUpper(r))
		}
		name = result.String()
	default:
		name = UpperFirst(CamelCase(name))
	}
	if len(name) > 0 && unicode.IsDigit(rune(name[0])) {
		name = "$" + name
	}
	return name
}

// ModelClass applies prefix/suffix to a model class name
func ModelClass(baseName string, opts *config.Options) string {
	return opts.ModelPrefix + TypeName(baseName, opts) + opts.ModelSuffix
}

// ServiceClass applies prefix/suffix to a service class name
func ServiceClass(baseName string, opts *config.Options) string {
	suffix := opts.ServiceSuffix
	if suffix == "" {
		suffix = "Service"
	}
	return opts.ServicePrefix + TypeName(baseName, opts) + suffix
}

// EnsureNotReserved adds $ suffix if name is a JS reserved keyword
func EnsureNotReserved(name string) string {
	if reservedKeywords[name] {
		return name + "$"
	}
	return name
}

// EscapeId returns a property/parameter name, quoted if not a valid JS identifier
func EscapeId(name string) string {
	matched, _ := regexp.MatchString(`^[a-zA-Z]\w*$`, name)
	if matched {
		return name
	}
	return "'" + strings.ReplaceAll(name, "'", "\\'") + "'"
}

// ToBasicChars converts text to basic letters/numbers/underscores
func ToBasicChars(text string, firstNonDigit bool) string {
	// deburr equivalent - strip diacritics
	text = strings.TrimSpace(text)
	// Remove non-word characters
	re := regexp.MustCompile(`[^\w$]+`)
	text = re.ReplaceAllString(text, "_")
	if firstNonDigit && len(text) > 0 && unicode.IsDigit(rune(text[0])) {
		text = "_" + text
	}
	return text
}

// UpperFirst capitalizes the first character
func UpperFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return string(unicode.ToUpper(r[0])) + string(r[1:])
}

// CamelCase converts a string to camelCase
func CamelCase(s string) string {
	// Split by underscore or space, lowercase first part, capitalize rest
	parts := splitIntoWords(s)
	if len(parts) == 0 {
		return ""
	}
	result := strings.ToLower(parts[0])
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += UpperFirst(strings.ToLower(parts[i]))
		}
	}
	return result
}

// KebabCase converts a string to kebab-case
func KebabCase(s string) string {
	parts := splitIntoWords(s)
	for i, p := range parts {
		parts[i] = strings.ToLower(p)
	}
	return strings.Join(parts, "-")
}

func splitIntoWords(s string) []string {
	// Split by underscores, spaces, or camelCase boundaries
	var words []string
	current := strings.Builder{}

	runes := []rune(s)
	for i, r := range runes {
		if r == '_' || r == ' ' || r == '-' {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			continue
		}
		if unicode.IsUpper(r) && i > 0 && (unicode.IsLower(rune(runes[i-1])) || (i+1 < len(runes) && unicode.IsLower(rune(runes[i+1])))) {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}

// PathToRoot returns a relative path from a given depth back to root
func PathToRoot(depth int) string {
	if depth <= 0 {
		return "./"
	}
	return strings.Repeat("../", depth)
}
