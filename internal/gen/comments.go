package gen

import (
	"strings"
)

// TsComments generates TypeScript JSDoc comments
func TsComments(description string, level int, deprecated ...bool) string {
	indent := strings.Repeat("  ", level)
	isDeprecated := len(deprecated) > 0 && deprecated[0]

	if description == "" {
		if isDeprecated {
			return indent + "/** @deprecated */"
		}
		return ""
	}

	lines := strings.Split(strings.TrimSpace(description), "\n")
	var sb strings.Builder
	sb.WriteString("\n" + indent + "/**\n")
	for _, line := range lines {
		if line == "" {
			sb.WriteString(indent + " *\n")
		} else {
			escaped := strings.ReplaceAll(line, "*/", "* /")
			sb.WriteString(indent + " * " + escaped + "\n")
		}
	}
	if isDeprecated {
		sb.WriteString(indent + " *\n")
		sb.WriteString(indent + " * @deprecated\n")
	}
	sb.WriteString(indent + " */\n" + indent)
	return sb.String()
}
