package gql_tag_operations

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
	"github.com/jzeiders/graphql-go-gen/pkg/plugins/base"
	"github.com/vektah/gqlparser/v2/ast"
)

// OperationOrFragment represents an operation or fragment with its name
type OperationOrFragment struct {
	InitialName string
	Definition  ast.Definition
}

// SourceWithOperations represents a source with its operations
type SourceWithOperations struct {
	Source     string
	Operations []OperationOrFragment
}

// Plugin generates the gql tag operations functionality
type Plugin struct{}

// New creates a new gql-tag-operations plugin
func New() plugin.Plugin {
	return &Plugin{}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "gql-tag-operations"
}

// Description returns the plugin description
func (p *Plugin) Description() string {
	return "Generates runtime graphql() function with operation lookup for TypeScript"
}

// DefaultConfig returns the default configuration
func (p *Plugin) DefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"gqlTagName":               "graphql",
		"useTypeImports":           false,
		"augmentedModuleName":      nil,
		"emitLegacyCommonJSImports": false,
		"documentMode":             "graphQLTag",
	}
}

// ValidateConfig validates the plugin configuration
func (p *Plugin) ValidateConfig(config map[string]interface{}) error {
	// Validate documentMode if provided
	if mode, ok := config["documentMode"].(string); ok {
		validModes := map[string]bool{
			"graphQLTag": true,
			"string":     true,
		}
		if !validModes[mode] {
			return fmt.Errorf("invalid documentMode: %s", mode)
		}
	}
	return nil
}

// Generate generates the gql tag operations code
func (p *Plugin) Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	// Get configuration
	gqlTagName := base.GetString(req.Config, "gqlTagName", "graphql")
	useTypeImports := base.GetBool(req.Config, "useTypeImports", false)
	augmentedModuleName := base.GetStringPtr(req.Config, "augmentedModuleName")
	emitLegacyCommonJSImports := base.GetBool(req.Config, "emitLegacyCommonJSImports", false)
	documentMode := base.GetString(req.Config, "documentMode", "graphQLTag")

	// Process sources from config
	sourcesWithOperations := p.processSources(req)

	// Special handling if sourcesWithOperations is provided in config
	if configSources, ok := req.Config["sourcesWithOperations"]; ok {
		if sources, ok := configSources.([]SourceWithOperations); ok {
			sourcesWithOperations = sources
		}
	}

	var sb strings.Builder

	// Generate based on document mode
	if documentMode == "string" {
		p.generateStringMode(&sb, sourcesWithOperations, gqlTagName, emitLegacyCommonJSImports)
	} else if augmentedModuleName != nil {
		p.generateAugmentedMode(&sb, sourcesWithOperations, gqlTagName, *augmentedModuleName, emitLegacyCommonJSImports)
	} else {
		p.generateStandardMode(&sb, sourcesWithOperations, gqlTagName, useTypeImports, emitLegacyCommonJSImports)
	}

	return &plugin.GenerateResponse{
		Files: map[string][]byte{
			req.OutputPath: []byte(sb.String()),
		},
	}, nil
}

// processSources processes documents to extract operations and fragments
func (p *Plugin) processSources(req *plugin.GenerateRequest) []SourceWithOperations {
	var result []SourceWithOperations

	// Group documents by source
	sourceMap := make(map[string][]OperationOrFragment)

	for _, doc := range req.Documents {
		if doc.Document == nil {
			continue
		}

		for _, def := range doc.Document.Definitions {
			var opOrFrag *OperationOrFragment

			switch d := def.(type) {
			case *ast.OperationDefinition:
				if d.Name == "" {
					// Skip anonymous operations with warning
					fmt.Printf("[client-preset] warning: anonymous operation skipped: %s\n", doc.Source)
					continue
				}
				opOrFrag = &OperationOrFragment{
					InitialName: p.getOperationVariableName(d),
					Definition:  d,
				}
			case *ast.FragmentDefinition:
				opOrFrag = &OperationOrFragment{
					InitialName: p.getFragmentVariableName(d),
					Definition:  d,
				}
			}

			if opOrFrag != nil {
				// Normalize linebreaks in source (CRLF to LF)
				normalizedSource := strings.ReplaceAll(doc.Source, "\r\n", "\n")
				sourceMap[normalizedSource] = append(sourceMap[normalizedSource], *opOrFrag)
			}
		}
	}

	// Convert map to slice
	for source, ops := range sourceMap {
		if len(ops) > 0 {
			// Take the first operation as the primary one
			result = append(result, SourceWithOperations{
				Source:     source,
				Operations: ops,
			})
		}
	}

	// Sort for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Source < result[j].Source
	})

	return result
}

// getOperationVariableName generates the variable name for an operation
func (p *Plugin) getOperationVariableName(op *ast.OperationDefinition) string {
	if op.Name == "" {
		return ""
	}
	// Convert operation name to PascalCase and add Document suffix
	return toPascalCase(op.Name) + "Document"
}

// getFragmentVariableName generates the variable name for a fragment
func (p *Plugin) getFragmentVariableName(frag *ast.FragmentDefinition) string {
	// Convert fragment name to PascalCase and add FragmentDoc suffix
	return toPascalCase(frag.Name) + "FragmentDoc"
}

// generateStringMode generates code for string document mode
func (p *Plugin) generateStringMode(sb *strings.Builder, sources []SourceWithOperations, gqlTagName string, emitLegacyCommonJSImports bool) {
	jsExt := ""
	if !emitLegacyCommonJSImports {
		jsExt = ".js"
	}

	sb.WriteString(fmt.Sprintf("import * as types from './graphql%s';\n\n", jsExt))

	// Generate document registry
	if len(sources) > 0 {
		p.generateDocumentRegistry(sb, sources, "augmented")
	} else {
		sb.WriteString("const documents = {};\n")
	}

	// Generate gql function overloads
	if len(sources) > 0 {
		p.generateGqlOverloads(sb, sources, gqlTagName, "augmented", emitLegacyCommonJSImports)
		sb.WriteString("\n")
	}

	// Generate main gql function
	sb.WriteString(fmt.Sprintf("export function %s(source: string) {\n", gqlTagName))
	sb.WriteString("  return (documents as any)[source] ?? {};\n")
	sb.WriteString("}\n")
}

// generateStandardMode generates code for standard mode with TypedDocumentNode
func (p *Plugin) generateStandardMode(sb *strings.Builder, sources []SourceWithOperations, gqlTagName string, useTypeImports bool, emitLegacyCommonJSImports bool) {
	jsExt := ""
	if !emitLegacyCommonJSImports {
		jsExt = ".js"
	}

	// Imports
	sb.WriteString(fmt.Sprintf("import * as types from './graphql%s';\n", jsExt))

	importType := "import"
	if useTypeImports {
		importType = "import type"
	}
	sb.WriteString(fmt.Sprintf("%s { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';\n\n", importType))

	// Generate document registry
	if len(sources) > 0 {
		p.generateDocumentRegistry(sb, sources, "lookup")
	} else {
		sb.WriteString("const documents = [];\n")
	}

	// JSDoc comment for the main function
	sb.WriteString("\n/**\n")
	sb.WriteString(fmt.Sprintf(" * The %s function is used to parse GraphQL queries into a document that can be used by GraphQL clients.\n", gqlTagName))
	sb.WriteString(" *\n")
	sb.WriteString(" * @example\n")
	sb.WriteString(" * ```ts\n")
	sb.WriteString(fmt.Sprintf(" * const query = %s(`query GetUser($id: ID!) { user(id: $id) { name } }`);\n", gqlTagName))
	sb.WriteString(" * ```\n")
	sb.WriteString(" *\n")
	sb.WriteString(" * The query argument is unknown!\n")
	sb.WriteString(" * Please regenerate the types.\n")
	sb.WriteString(" */\n")
	sb.WriteString(fmt.Sprintf("export function %s(source: string): unknown;\n\n", gqlTagName))

	// Generate gql function overloads
	if len(sources) > 0 {
		p.generateGqlOverloads(sb, sources, gqlTagName, "lookup", emitLegacyCommonJSImports)
		sb.WriteString("\n")
	}

	// Main gql function implementation
	sb.WriteString(fmt.Sprintf("export function %s(source: string) {\n", gqlTagName))
	sb.WriteString("  return (documents as any)[source] ?? {};\n")
	sb.WriteString("}\n\n")

	// DocumentType helper
	sb.WriteString("export type DocumentType<TDocumentNode extends DocumentNode<any, any>> = TDocumentNode extends DocumentNode<\n")
	sb.WriteString("  infer TType,\n")
	sb.WriteString("  any\n")
	sb.WriteString(">\n")
	sb.WriteString("  ? TType\n")
	sb.WriteString("  : never;\n")
}

// generateAugmentedMode generates code for module augmentation mode
func (p *Plugin) generateAugmentedMode(sb *strings.Builder, sources []SourceWithOperations, gqlTagName string, augmentedModuleName string, emitLegacyCommonJSImports bool) {
	sb.WriteString("import { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';\n")
	sb.WriteString(fmt.Sprintf("declare module \"%s\" {\n", augmentedModuleName))

	// Indent content
	var content strings.Builder
	content.WriteString("\n")

	if len(sources) > 0 {
		p.generateGqlOverloads(&content, sources, gqlTagName, "augmented", emitLegacyCommonJSImports)
	}

	content.WriteString(fmt.Sprintf("export function %s(source: string): unknown;\n\n", gqlTagName))

	// DocumentType helper
	content.WriteString("export type DocumentType<TDocumentNode extends DocumentNode<any, any>> = TDocumentNode extends DocumentNode<\n")
	content.WriteString("  infer TType,\n")
	content.WriteString("  any\n")
	content.WriteString(">\n")
	content.WriteString("  ? TType\n")
	content.WriteString("  : never;\n")

	// Indent all lines
	lines := strings.Split(content.String(), "\n")
	for _, line := range lines {
		if line == "" {
			sb.WriteString("\n")
		} else {
			sb.WriteString("  " + line + "\n")
		}
	}

	sb.WriteString("}\n")
}

// generateDocumentRegistry generates the document registry
func (p *Plugin) generateDocumentRegistry(sb *strings.Builder, sources []SourceWithOperations, mode string) {
	sb.WriteString("/**\n")
	sb.WriteString(" * Map of all GraphQL operations in the project.\n")
	sb.WriteString(" *\n")
	sb.WriteString(" * This map has several performance disadvantages:\n")
	sb.WriteString(" * 1. It is not tree-shakeable, so it will include all operations in the project.\n")
	sb.WriteString(" * 2. It is not minifiable, so the string of a GraphQL query will be multiple times inside the bundle.\n")
	sb.WriteString(" * 3. It does not support dead code elimination, so it will add unused operations.\n")
	sb.WriteString(" *\n")
	sb.WriteString(" * Therefore it is highly recommended to use the babel or swc plugin for production.\n")
	sb.WriteString(" * Learn more about it here: https://the-guild.dev/graphql/codegen/plugins/presets/preset-client#reducing-bundle-size\n")
	sb.WriteString(" */\n")

	// Use a map to dedupe
	seenLines := make(map[string]bool)

	// Type definition
	sb.WriteString("type Documents = {\n")
	for _, source := range sources {
		if len(source.Operations) > 0 {
			line := fmt.Sprintf("    %s: typeof types.%s,\n", escapeString(source.Source), source.Operations[0].InitialName)
			if !seenLines[line] {
				sb.WriteString(line)
				seenLines[line] = true
			}
		}
	}
	sb.WriteString("};\n")

	// Reset for actual values
	seenLines = make(map[string]bool)

	// Actual document registry
	sb.WriteString("const documents: Documents = {\n")
	for _, source := range sources {
		if len(source.Operations) > 0 {
			line := fmt.Sprintf("    %s: types.%s,\n", escapeString(source.Source), source.Operations[0].InitialName)
			if !seenLines[line] {
				sb.WriteString(line)
				seenLines[line] = true
			}
		}
	}
	sb.WriteString("};\n")
}

// generateGqlOverloads generates the overloaded gql function signatures
func (p *Plugin) generateGqlOverloads(sb *strings.Builder, sources []SourceWithOperations, gqlTagName string, mode string, emitLegacyCommonJSImports bool) {
	// Use a set to dedupe
	seen := make(map[string]bool)

	for _, source := range sources {
		if len(source.Operations) == 0 {
			continue
		}

		var returnType string
		if mode == "lookup" {
			returnType = fmt.Sprintf("(typeof documents)[%s]", escapeString(source.Source))
		} else {
			jsExt := ""
			if !emitLegacyCommonJSImports {
				jsExt = ".js"
			}
			returnType = fmt.Sprintf("typeof import('./graphql%s').%s", jsExt, source.Operations[0].InitialName)
		}

		signature := fmt.Sprintf("/**\n * The %s function is used to parse GraphQL queries into a document that can be used by GraphQL clients.\n */\nexport function %s(source: %s): %s;\n",
			gqlTagName, gqlTagName, escapeString(source.Source), returnType)

		if !seen[signature] {
			sb.WriteString(signature)
			seen[signature] = true
		}
	}
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	// Simple implementation - first letter uppercase
	return strings.ToUpper(s[:1]) + s[1:]
}

// escapeString escapes a string for use in TypeScript
func escapeString(s string) string {
	// Use JSON marshaling for proper escaping
	escaped := fmt.Sprintf("%q", s)
	// Replace backticks if any
	escaped = strings.ReplaceAll(escaped, "`", "\\`")
	return escaped
}