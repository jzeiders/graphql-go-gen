package client

import (
	"fmt"
	"strings"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/vektah/gqlparser/v2/ast"
)

// OperationOrFragment represents an operation or fragment with its generated name
type OperationOrFragment struct {
	InitialName string
	Operation   *ast.OperationDefinition
	Fragment    *ast.FragmentDefinition
}

// SourceWithOperations represents a source document with its operations
type SourceWithOperations struct {
	Source     string
	Operations []OperationOrFragment
}

// BuildNameFunction is a function that generates variable names for operations/fragments
type BuildNameFunction func(op *ast.OperationDefinition, frag *ast.FragmentDefinition) string

// ProcessSources processes GraphQL documents to extract operations and fragments
func ProcessSources(docs []*documents.Document, buildName BuildNameFunction) []SourceWithOperations {
	var result []SourceWithOperations

	for _, doc := range docs {
		if doc.AST == nil {
			continue
		}

		var operations []OperationOrFragment

		// Process operations
		for _, op := range doc.AST.Operations {
			if op.Name == "" {
				// Log warning for anonymous operations
				fmt.Printf("[client-preset] warning: anonymous operation skipped: %s\n", doc.Content)
				continue
			}
			operations = append(operations, OperationOrFragment{
				InitialName: buildName(op, nil),
				Operation:   op,
			})
		}

		// Process fragments
		for _, frag := range doc.AST.Fragments {
			operations = append(operations, OperationOrFragment{
				InitialName: buildName(nil, frag),
				Fragment:    frag,
			})
		}

		if len(operations) > 0 {
			// Normalize linebreaks (CRLF to LF) for cross-platform compatibility
			normalizedSource := fixLinebreaks(doc.Content)

			result = append(result, SourceWithOperations{
				Source:     normalizedSource,
				Operations: operations,
			})
		}
	}

	return result
}

// fixLinebreaks normalizes linebreaks from CRLF to LF
// This ensures consistent string comparison across platforms (Windows vs Unix)
// JavaScript/TypeScript template literals always use LF regardless of OS
func fixLinebreaks(source string) string {
	return strings.ReplaceAll(source, "\r\n", "\n")
}

// DefaultBuildName generates default variable names for operations and fragments
func DefaultBuildName(op *ast.OperationDefinition, frag *ast.FragmentDefinition) string {
	if op != nil {
		return getOperationVariableName(op)
	}
	if frag != nil {
		return getFragmentVariableName(frag)
	}
	return ""
}

// getOperationVariableName generates the variable name for an operation
func getOperationVariableName(op *ast.OperationDefinition) string {
	if op.Name == "" {
		return ""
	}
	// Convert to PascalCase and add Document suffix
	return toPascalCase(op.Name) + "Document"
}

// getFragmentVariableName generates the variable name for a fragment
func getFragmentVariableName(frag *ast.FragmentDefinition) string {
	// Convert to PascalCase and add FragmentDoc suffix
	return toPascalCase(frag.Name) + "FragmentDoc"
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	// Simple implementation - uppercase first letter
	return strings.ToUpper(s[:1]) + s[1:]
}