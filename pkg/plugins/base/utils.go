package base

import (
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

// GetBool safely gets a boolean value from a map
func GetBool(m map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// GetString safely gets a string value from a map
func GetString(m map[string]interface{}, key string, defaultValue string) string {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return defaultValue
}

// TypeToTypeScript converts a GraphQL type to TypeScript
func TypeToTypeScript(t *ast.Type, scalarMap map[string]string, strictNulls bool) string {
	if t == nil {
		return "unknown"
	}

	var result string

	// Handle the base type
	switch t.NamedType {
	case "String", "ID":
		result = "string"
	case "Int", "Float":
		result = "number"
	case "Boolean":
		result = "boolean"
	default:
		// Check scalar map for custom scalar mappings
		if mapped, ok := scalarMap[t.NamedType]; ok {
			result = mapped
		} else {
			result = t.NamedType
		}
	}

	// Handle list types
	if t.Elem != nil {
		result = TypeToTypeScript(t.Elem, scalarMap, strictNulls) + "[]"
	}

	// Handle nullability
	if !t.NonNull && strictNulls {
		result = result + " | null"
	}

	return result
}

// FormatComment formats a GraphQL description as a TypeScript comment
func FormatComment(description string, indent string) string {
	if description == "" {
		return ""
	}

	lines := strings.Split(description, "\n")
	if len(lines) == 1 {
		return indent + "/** " + strings.TrimSpace(description) + " */\n"
	}

	var sb strings.Builder
	sb.WriteString(indent + "/**\n")
	for _, line := range lines {
		sb.WriteString(indent + " * " + strings.TrimSpace(line) + "\n")
	}
	sb.WriteString(indent + " */\n")
	return sb.String()
}

// ToPascalCase converts a string to PascalCase
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}
	// Simple implementation - can be enhanced
	return strings.ToUpper(s[:1]) + s[1:]
}

// ToCamelCase converts a string to camelCase
func ToCamelCase(s string) string {
	if s == "" {
		return ""
	}
	// Simple implementation - can be enhanced
	return strings.ToLower(s[:1]) + s[1:]
}