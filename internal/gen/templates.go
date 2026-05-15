package gen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// TemplateManager manages Go templates for code generation
type TemplateManager struct {
	templates map[string]*template.Template
	funcs     template.FuncMap
}

// NewTemplateManager creates a new template manager with built-in and optional custom templates
func NewTemplateManager(builtInDir string, customDir string) (*TemplateManager, error) {
	tm := &TemplateManager{
		templates: make(map[string]*template.Template),
		funcs: template.FuncMap{
			"upperFirst": UpperFirst,
			"lowerFirst": func(s string) string {
				if s == "" {
					return s
				}
				return strings.ToLower(s[:1]) + s[1:]
			},
			"camelCase":  CamelCase,
			"kebabCase":  KebabCase,
			"fileName":   FileName,
			"typeName":   func(s string) string { return TypeName(s, nil) },
			"escapeJS":   EscapeJS,
			"escapeId":   EscapeId,
			"tsComments": TsComments,
			"join":       strings.Join,
			"add":        func(a, b int) int { return a + b },
			"seq": func(n int) []int {
				r := make([]int, n)
				for i := 0; i < n; i++ {
					r[i] = i
				}
				return r
			},
			"dict": func(values ...any) (map[string]any, error) {
				if len(values)%2 != 0 {
					return nil, fmt.Errorf("dict expects even number of args")
				}
				d := make(map[string]any)
				for i := 0; i < len(values); i += 2 {
					key, ok := values[i].(string)
					if !ok {
						return nil, fmt.Errorf("dict keys must be strings")
					}
					d[key] = values[i+1]
				}
				return d, nil
			},
		},
	}

	// Load built-in templates
	if err := tm.loadDir(builtInDir); err != nil {
		return nil, fmt.Errorf("load built-in templates: %w", err)
	}

	// Load custom templates if directory exists
	if customDir != "" {
		if err := tm.loadDir(customDir); err != nil {
			return nil, fmt.Errorf("load custom templates: %w", err)
		}
	}

	return tm, nil
}

func (tm *TemplateManager) loadDir(dir string) error {
	entries, err := filepath.Glob(filepath.Join(dir, "*.go.tmpl"))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := filepath.Base(entry)
		baseName := strings.TrimSuffix(name, ".go.tmpl")
		data, err := os.ReadFile(entry)
		if err != nil {
			return fmt.Errorf("read template %s: %w", name, err)
		}
		t := template.New(baseName).Funcs(tm.funcs)
		t, err = t.Parse(string(data))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", name, err)
		}
		tm.templates[baseName] = t
	}
	return nil
}

// Apply renders a template with the given data
func (tm *TemplateManager) Apply(name string, data any) (string, error) {
	t, ok := tm.templates[name]
	if !ok {
		return "", fmt.Errorf("template not found: %s", name)
	}
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %s: %w", name, err)
	}
	return buf.String(), nil
}

// HasTemplate returns true if a template with the given name exists
func (tm *TemplateManager) HasTemplate(name string) bool {
	_, ok := tm.templates[name]
	return ok
}
