package fragment_masking

import (
	"bytes"
	"fmt"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
	"github.com/vektah/gqlparser/v2/ast"
)

// Plugin generates fragment masking utilities for TypeScript
type Plugin struct{}

// Config for the fragment-masking plugin
type Config struct {
	// UnmaskFunctionName is the name of the function used to unmask fragments (default: "useFragment")
	UnmaskFunctionName string `yaml:"unmaskFunctionName" json:"unmaskFunctionName"`
	// UseTypeImports controls whether to use TypeScript type imports
	UseTypeImports bool `yaml:"useTypeImports" json:"useTypeImports"`
	// EmitLegacyCommonJSImports controls whether to emit legacy CommonJS imports
	EmitLegacyCommonJSImports bool `yaml:"emitLegacyCommonJSImports" json:"emitLegacyCommonJSImports"`
	// IsStringDocumentMode indicates if documents are in string mode
	IsStringDocumentMode bool `yaml:"isStringDocumentMode" json:"isStringDocumentMode"`
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "fragment-masking"
}

// Generate generates the fragment masking utilities
func (p *Plugin) Generate(schema *ast.Schema, documents []*documents.Document, cfg interface{}) ([]byte, error) {
	config := p.parseConfig(cfg)

	var buf bytes.Buffer

	// Generate header
	buf.WriteString("/* eslint-disable */\n")

	// Generate imports
	if config.UseTypeImports {
		buf.WriteString("import type { ResultOf, DocumentTypeDecoration, TypedDocumentNode } from '@graphql-typed-document-node/core';\n")
		buf.WriteString("import type { FragmentDefinitionNode } from 'graphql';\n")
		buf.WriteString("import type { Incremental } from './graphql';\n\n")
	} else {
		buf.WriteString("import { ResultOf, DocumentTypeDecoration, TypedDocumentNode } from '@graphql-typed-document-node/core';\n")
		buf.WriteString("import { FragmentDefinitionNode } from 'graphql';\n")
		buf.WriteString("import { Incremental } from './graphql';\n\n")
	}

	// Generate utility types
	buf.WriteString("export type FragmentType<TDocumentType extends DocumentTypeDecoration<any, any>> =\n")
	buf.WriteString("  TDocumentType extends DocumentTypeDecoration<infer TType, any> ? TType : never;\n\n")

	// Generate makeFragmentData function
	unmaskName := config.UnmaskFunctionName
	if unmaskName == "" {
		unmaskName = "useFragment"
	}

	buf.WriteString("export function makeFragmentData<\n")
	buf.WriteString("  F extends DocumentTypeDecoration<any, any>,\n")
	buf.WriteString("  FT extends ResultOf<F>\n")
	buf.WriteString(">(data: FT, _fragment: F): FragmentType<F> {\n")
	buf.WriteString("  return data as FragmentType<F>;\n")
	buf.WriteString("}\n\n")

	// Generate useFragment function
	buf.WriteString(fmt.Sprintf("export function %s<TType>(\n", unmaskName))
	buf.WriteString("  _documentNode: DocumentTypeDecoration<TType, any>,\n")
	buf.WriteString("  fragmentType: FragmentType<DocumentTypeDecoration<TType, any>>\n")
	buf.WriteString("): TType;\n")

	// Generate overload for array
	buf.WriteString(fmt.Sprintf("export function %s<TType>(\n", unmaskName))
	buf.WriteString("  _documentNode: DocumentTypeDecoration<TType, any>,\n")
	buf.WriteString("  fragmentType: Array<FragmentType<DocumentTypeDecoration<TType, any>>>\n")
	buf.WriteString("): Array<TType>;\n")

	// Generate overload for nullable
	buf.WriteString(fmt.Sprintf("export function %s<TType>(\n", unmaskName))
	buf.WriteString("  _documentNode: DocumentTypeDecoration<TType, any>,\n")
	buf.WriteString("  fragmentType: FragmentType<DocumentTypeDecoration<TType, any>> | null | undefined\n")
	buf.WriteString("): TType | null | undefined;\n")

	// Generate overload for nullable array
	buf.WriteString(fmt.Sprintf("export function %s<TType>(\n", unmaskName))
	buf.WriteString("  _documentNode: DocumentTypeDecoration<TType, any>,\n")
	buf.WriteString("  fragmentType: Array<FragmentType<DocumentTypeDecoration<TType, any>>> | null | undefined\n")
	buf.WriteString("): Array<TType> | null | undefined;\n")

	// Generate implementation
	buf.WriteString(fmt.Sprintf("export function %s<TType>(\n", unmaskName))
	buf.WriteString("  _documentNode: DocumentTypeDecoration<TType, any>,\n")
	buf.WriteString("  fragmentType:\n")
	buf.WriteString("    | FragmentType<DocumentTypeDecoration<TType, any>>\n")
	buf.WriteString("    | Array<FragmentType<DocumentTypeDecoration<TType, any>>>\n")
	buf.WriteString("    | null\n")
	buf.WriteString("    | undefined\n")
	buf.WriteString("): TType | Array<TType> | null | undefined {\n")
	buf.WriteString("  return fragmentType as any;\n")
	buf.WriteString("}\n\n")

	// Generate makeFragmentData helper
	buf.WriteString("export function makeFragmentData<\n")
	buf.WriteString("  F extends DocumentTypeDecoration<any, any>,\n")
	buf.WriteString("  FT extends ResultOf<F>\n")
	buf.WriteString(">(data: FT, _fragment: F): FragmentType<F> {\n")
	buf.WriteString("  return data as FragmentType<F>;\n")
	buf.WriteString("}\n\n")

	// Generate isFragmentReady for deferred/incremental support
	buf.WriteString("export function isFragmentReady<TQuery, TFrag>(\n")
	buf.WriteString("  queryNode: DocumentTypeDecoration<TQuery, any>,\n")
	buf.WriteString("  fragmentNode: TypedDocumentNode<TFrag>,\n")
	buf.WriteString("  data: FragmentType<DocumentTypeDecoration<Incremental<TQuery>, any>> | null | undefined\n")
	buf.WriteString("): data is FragmentType<DocumentTypeDecoration<TQuery, any>> {\n")

	if config.IsStringDocumentMode {
		// String mode implementation
		buf.WriteString("  const deferredFields = queryNode.definitions\n")
		buf.WriteString("    .filter((definition) => definition.kind === 'FragmentDefinition' && definition.directives?.some((directive) => directive.name.value === 'defer'))\n")
		buf.WriteString("    .map((definition) => (definition as FragmentDefinitionNode).name.value);\n\n")
		buf.WriteString("  const fragName = fragmentNode.definitions[0]?.name?.value;\n\n")
		buf.WriteString("  return !deferredFields.includes(fragName) && data != null;\n")
	} else {
		// Document mode implementation
		buf.WriteString("  const deferredFields = (queryNode as { __meta__?: { deferredFields: Record<string, (keyof TFrag)[]> } })\n")
		buf.WriteString("    .__meta__?.deferredFields;\n\n")
		buf.WriteString("  const fragDef = fragmentNode.definitions[0] as FragmentDefinitionNode | undefined;\n")
		buf.WriteString("  const fragName = fragDef?.name?.value;\n\n")
		buf.WriteString("  const fields = (fragName && deferredFields?.[fragName]) || [];\n")
		buf.WriteString("  return fields.length > 0 && fields.every(field => data && field in data);\n")
	}

	buf.WriteString("}\n")

	return buf.Bytes(), nil
}

// parseConfig parses the plugin configuration
func (p *Plugin) parseConfig(cfg interface{}) *Config {
	config := &Config{}
	if cfg == nil {
		return config
	}

	if mapConfig, ok := cfg.(map[string]interface{}); ok {
		if unmaskName, ok := mapConfig["unmaskFunctionName"].(string); ok {
			config.UnmaskFunctionName = unmaskName
		}
		if useTypeImports, ok := mapConfig["useTypeImports"].(bool); ok {
			config.UseTypeImports = useTypeImports
		}
		if emitLegacy, ok := mapConfig["emitLegacyCommonJSImports"].(bool); ok {
			config.EmitLegacyCommonJSImports = emitLegacy
		}
		if isStringMode, ok := mapConfig["isStringDocumentMode"].(bool); ok {
			config.IsStringDocumentMode = isStringMode
		}
	}

	return config
}

// Register registers the plugin
func init() {
	plugin.Register("fragment-masking", &Plugin{})
}