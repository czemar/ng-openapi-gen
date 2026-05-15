package ngopenapigen

import "embed"

//go:embed templates/*.go.tmpl
var TemplatesFS embed.FS
