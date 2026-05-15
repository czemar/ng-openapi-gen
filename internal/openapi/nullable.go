package openapi

// IsNullableSchema checks if a schema is nullable (compatible with 3.0 and 3.1)
func IsNullableSchema(schema *Schema) bool {
	if schema.Nullable {
		return true
	}
	if schema.Type != nil {
		if types, ok := schema.Type.([]any); ok {
			for _, t := range types {
				if s, ok := t.(string); ok && s == "null" {
					return true
				}
			}
		}
	}
	return false
}
