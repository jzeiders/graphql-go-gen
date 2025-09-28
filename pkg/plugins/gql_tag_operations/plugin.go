package gql_tag_operations

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
	"github.com/vektah/gqlparser/v2/ast"
)

// Plugin generates TypeScript gql tag functions with type-safe overloads
type Plugin struct{}

// Config for the gql-tag-operations plugin
type Config struct {
	// GqlTagName is the name of the GraphQL tag function (default: "graphql")
	GqlTagName string `yaml:"gqlTagName" json:"gqlTagName"`
	// SourcesWithOperations contains the operations grouped by source
	SourcesWithOperations []SourceWithOperations `yaml:"sourcesWithOperations" json:"sourcesWithOperations"`
}

// SourceWithOperations represents a source file with its operations
type SourceWithOperations struct {
	Source     *documents.Document
	Operations []OperationDescriptor
}

// OperationDescriptor describes a GraphQL operation
type OperationDescriptor struct {
	Name         string
	VariableName string
	Content      string
	Type         ast.Operation
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "gql-tag-operations"
}

// Generate generates the gql tag operations code
func (p *Plugin) Generate(schema *ast.Schema, documents []*documents.Document, cfg interface{}) ([]byte, error) {
	config := p.parseConfig(cfg)

	var buf bytes.Buffer
	buf.WriteString("/* eslint-disable */\n")
	buf.WriteString("import * as types from './graphql';\n")
	buf.WriteString("import type { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';\n\n")

	// Generate comment about bundle size optimization
	buf.WriteString("/**\n")
	buf.WriteString(" * Map of all GraphQL operations in the project.\n")
	buf.WriteString(" *\n")
	buf.WriteString(" * This map has several performance disadvantages:\n")
	buf.WriteString(" * 1. It is not tree-shakeable, so it will include all operations in the project.\n")
	buf.WriteString(" * 2. It is not minifiable, so the string of a GraphQL query will be multiple times inside the bundle.\n")
	buf.WriteString(" * 3. It does not support dead code elimination, so it will add unused operations.\n")
	buf.WriteString(" *\n")
	buf.WriteString(" * Therefore it is highly recommended to use the babel or swc plugin for production.\n")
	buf.WriteString(" */\n")

	// Collect all operations
	operations := p.collectOperations(documents)

	// Generate the Documents type
	buf.WriteString("type Documents = {\n")
	for _, op := range operations {
		buf.WriteString(fmt.Sprintf("  '%s': typeof types.%sDocument;\n",
			p.escapeString(op.Content), op.VariableName))
	}
	buf.WriteString("};\n")

	// Generate the documents object
	buf.WriteString("const documents: Documents = {\n")
	for _, op := range operations {
		buf.WriteString(fmt.Sprintf("  '%s': types.%sDocument,\n",
			p.escapeString(op.Content), op.VariableName))
	}
	buf.WriteString("};\n\n")

	// Generate the graphql function overloads
	tagName := config.GqlTagName
	if tagName == "" {
		tagName = "graphql"
	}

	// Default overload for unknown strings
	buf.WriteString("/**\n")
	buf.WriteString(fmt.Sprintf(" * The %s function is used to parse GraphQL queries into a document that can be used by GraphQL clients.\n", tagName))
	buf.WriteString(" *\n")
	buf.WriteString(" *\n")
	buf.WriteString(" * @example\n")
	buf.WriteString(" * ```ts\n")
	buf.WriteString(fmt.Sprintf(" * const query = %s(`query GetUser($id: ID!) { user(id: $id) { name } }`);\n", tagName))
	buf.WriteString(" * ```\n")
	buf.WriteString(" *\n")
	buf.WriteString(" * The query argument is unknown!\n")
	buf.WriteString(" * Please regenerate the types.\n")
	buf.WriteString(" */\n")
	buf.WriteString(fmt.Sprintf("export function %s(source: string): unknown;\n\n", tagName))

	// Generate specific overloads for each operation
	for _, op := range operations {
		buf.WriteString("/**\n")
		buf.WriteString(fmt.Sprintf(" * The %s function is used to parse GraphQL queries into a document that can be used by GraphQL clients.\n", tagName))
		buf.WriteString(" */\n")
		buf.WriteString(fmt.Sprintf("export function %s(source: '%s'): (typeof documents)['%s'];\n",
			tagName, p.escapeString(op.Content), p.escapeString(op.Content)))
	}

	// Generate the implementation
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("export function %s(source: string) {\n", tagName))
	buf.WriteString("  return (documents as any)[source] ?? {};\n")
	buf.WriteString("}\n\n")

	// Export the type
	buf.WriteString("export type DocumentType<TDocumentNode extends DocumentNode<any, any>> = TDocumentNode extends DocumentNode<\n")
	buf.WriteString("  infer TType,\n")
	buf.WriteString("  any\n")
	buf.WriteString(">\n")
	buf.WriteString("  ? TType\n")
	buf.WriteString("  : never;\n")

	return buf.Bytes(), nil
}

// parseConfig parses the plugin configuration
func (p *Plugin) parseConfig(cfg interface{}) *Config {
	config := &Config{}
	if cfg == nil {
		return config
	}

	if mapConfig, ok := cfg.(map[string]interface{}); ok {
		if tagName, ok := mapConfig["gqlTagName"].(string); ok {
			config.GqlTagName = tagName
		}
		if sources, ok := mapConfig["sourcesWithOperations"].([]SourceWithOperations); ok {
			config.SourcesWithOperations = sources
		}
	}

	return config
}

// collectOperations collects all operations from documents
func (p *Plugin) collectOperations(documents []*documents.Document) []OperationDescriptor {
	var operations []OperationDescriptor

	for _, doc := range documents {
		if doc.AST == nil {
			continue
		}

		// Process operations
		for _, op := range doc.AST.Operations {
				op := OperationDescriptor{
					Name:         def.Name,
					Type:         def.Operation,
					Content:      p.getOperationContent(doc, def),
					VariableName: p.getOperationVariableName(def),
				}
				operations = append(operations, op)
			case *ast.FragmentDefinition:
				op := OperationDescriptor{
					Name:         def.Name,
					Content:      p.getFragmentContent(doc, def),
					VariableName: p.getFragmentVariableName(def),
				}
				operations = append(operations, op)
			}
		}
	}

	return operations
}

// getOperationContent extracts the content of an operation
func (p *Plugin) getOperationContent(doc *documents.Document, op *ast.OperationDefinition) string {
	// Try to get the original content from the document
	if doc.Content != "" {
		return strings.TrimSpace(doc.Content)
	}

	// Fallback to reconstructing from AST
	return p.reconstructOperation(op)
}

// getFragmentContent extracts the content of a fragment
func (p *Plugin) getFragmentContent(doc *documents.Document, frag *ast.FragmentDefinition) string {
	// Try to get the original content from the document
	if doc.Content != "" {
		return strings.TrimSpace(doc.Content)
	}

	// Fallback to reconstructing from AST
	return p.reconstructFragment(frag)
}

// reconstructOperation reconstructs an operation from its AST
func (p *Plugin) reconstructOperation(op *ast.OperationDefinition) string {
	// This is a simplified reconstruction - in production you'd want to use a proper printer
	var buf bytes.Buffer

	buf.WriteString(string(op.Operation))
	if op.Name != "" {
		buf.WriteString(" ")
		buf.WriteString(op.Name)
	}

	if len(op.VariableDefinitions) > 0 {
		buf.WriteString("(")
		for i, v := range op.VariableDefinitions {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString("$")
			buf.WriteString(v.Variable)
			buf.WriteString(": ")
			buf.WriteString(v.Type.String())
		}
		buf.WriteString(")")
	}

	buf.WriteString(" { ... }")
	return buf.String()
}

// reconstructFragment reconstructs a fragment from its AST
func (p *Plugin) reconstructFragment(frag *ast.FragmentDefinition) string {
	// This is a simplified reconstruction
	return fmt.Sprintf("fragment %s on %s { ... }", frag.Name, frag.TypeCondition)
}

// getOperationVariableName generates a variable name for an operation
func (p *Plugin) getOperationVariableName(op *ast.OperationDefinition) string {
	if op.Name != "" {
		return p.capitalizeFirst(op.Name) + p.capitalizeFirst(string(op.Operation))
	}
	return p.capitalizeFirst(string(op.Operation))
}

// getFragmentVariableName generates a variable name for a fragment
func (p *Plugin) getFragmentVariableName(frag *ast.FragmentDefinition) string {
	return p.capitalizeFirst(frag.Name) + "FragmentDoc"
}

// capitalizeFirst capitalizes the first letter of a string
func (p *Plugin) capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// escapeString escapes a string for use in TypeScript
func (p *Plugin) escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// Register registers the plugin
func init() {
	plugin.Register("gql-tag-operations", &Plugin{})
}