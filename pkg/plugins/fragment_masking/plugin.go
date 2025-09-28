package fragment_masking

import (
	"context"
	"fmt"
	"strings"

	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
	"github.com/jzeiders/graphql-go-gen/pkg/plugins/base"
)

// Plugin generates fragment masking helper functions
type Plugin struct{}

// New creates a new fragment-masking plugin
func New() plugin.Plugin {
	return &Plugin{}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "fragment-masking"
}

// Description returns the plugin description
func (p *Plugin) Description() string {
	return "Generates fragment masking utilities for type-safe fragment composition"
}

// DefaultConfig returns the default configuration
func (p *Plugin) DefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"unmaskFunctionName":        "useFragment",
		"useTypeImports":            false,
		"augmentedModuleName":       nil,
		"emitLegacyCommonJSImports": false,
		"isStringDocumentMode":      false,
	}
}

// ValidateConfig validates the plugin configuration
func (p *Plugin) ValidateConfig(config map[string]interface{}) error {
	// All config options are optional
	return nil
}

// Generate generates fragment masking utilities
func (p *Plugin) Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	// Get configuration
	unmaskFunctionName := base.GetString(req.Config, "unmaskFunctionName", "useFragment")
	useTypeImports := base.GetBool(req.Config, "useTypeImports", false)
	augmentedModuleName := base.GetStringPtr(req.Config, "augmentedModuleName")
	emitLegacyCommonJSImports := base.GetBool(req.Config, "emitLegacyCommonJSImports", false)
	isStringDocumentMode := base.GetBool(req.Config, "isStringDocumentMode", false)

	var sb strings.Builder

	if augmentedModuleName != nil {
		p.generateAugmentedMode(&sb, unmaskFunctionName, useTypeImports, *augmentedModuleName)
	} else {
		p.generateStandardMode(&sb, unmaskFunctionName, useTypeImports, emitLegacyCommonJSImports, isStringDocumentMode)
	}

	return &plugin.GenerateResponse{
		Files: map[string][]byte{
			req.OutputPath: []byte(sb.String()),
		},
	}, nil
}

// generateStandardMode generates the standard fragment masking utilities
func (p *Plugin) generateStandardMode(sb *strings.Builder, unmaskFunctionName string, useTypeImports bool, emitLegacyCommonJSImports bool, isStringDocumentMode bool) {
	// Imports
	importType := "import"
	if useTypeImports {
		importType = "import type"
	}

	documentNodeImports := "ResultOf, DocumentTypeDecoration"
	if !isStringDocumentMode {
		documentNodeImports += ", TypedDocumentNode"
	}
	sb.WriteString(fmt.Sprintf("%s { %s } from '@graphql-typed-document-node/core';\n", importType, documentNodeImports))

	if !isStringDocumentMode {
		sb.WriteString(fmt.Sprintf("%s { FragmentDefinitionNode } from 'graphql';\n", importType))
	}

	jsExt := ""
	if !emitLegacyCommonJSImports {
		jsExt = ".js"
	}

	incrementalImports := "Incremental"
	if isStringDocumentMode {
		incrementalImports += ", TypedDocumentString"
	}
	sb.WriteString(fmt.Sprintf("%s { %s } from './graphql%s';\n\n", importType, incrementalImports, jsExt))

	// FragmentType helper
	p.writeFragmentTypeHelper(sb)
	sb.WriteString("\n")

	// Unmask function with all overloads
	p.writeUnmaskFunction(sb, unmaskFunctionName)
	sb.WriteString("\n")

	// makeFragmentData helper
	p.writeMakeFragmentDataHelper(sb)
	sb.WriteString("\n")

	// isFragmentReady helper
	p.writeIsFragmentReadyFunction(sb, isStringDocumentMode)
}

// generateAugmentedMode generates module augmentation mode
func (p *Plugin) generateAugmentedMode(sb *strings.Builder, unmaskFunctionName string, useTypeImports bool, augmentedModuleName string) {
	importType := "import"
	if useTypeImports {
		importType = "import type"
	}

	sb.WriteString(fmt.Sprintf("%s { ResultOf, DocumentTypeDecoration } from '@graphql-typed-document-node/core';\n", importType))
	sb.WriteString(fmt.Sprintf("declare module \"%s\" {\n", augmentedModuleName))

	// Indent all content
	var content strings.Builder

	// FragmentType helper (indented)
	p.writeFragmentTypeHelper(&content)
	content.WriteString("\n")

	// Unmask function type definitions only (indented)
	p.writeUnmaskFunctionTypeDefinitions(&content, unmaskFunctionName)
	content.WriteString("\n")

	// makeFragmentData helper (indented)
	p.writeMakeFragmentDataHelper(&content)

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

// writeFragmentTypeHelper writes the FragmentType type helper
func (p *Plugin) writeFragmentTypeHelper(sb *strings.Builder) {
	sb.WriteString("export type FragmentType<TDocumentType extends DocumentTypeDecoration<any, any>> = TDocumentType extends DocumentTypeDecoration<\n")
	sb.WriteString("  infer TType,\n")
	sb.WriteString("  any\n")
	sb.WriteString(">\n")
	sb.WriteString("  ? [TType] extends [{ ' $fragmentName'?: infer TKey }]\n")
	sb.WriteString("    ? TKey extends string\n")
	sb.WriteString("      ? { ' $fragmentRefs'?: { [key in TKey]: TType } }\n")
	sb.WriteString("      : never\n")
	sb.WriteString("    : never\n")
	sb.WriteString("  : never;")
}

// writeUnmaskFunctionTypeDefinitions writes just the type definitions for the unmask function
func (p *Plugin) writeUnmaskFunctionTypeDefinitions(sb *strings.Builder, unmaskFunctionName string) {
	// Non-nullable overload
	sb.WriteString("// return non-nullable if `fragmentType` is non-nullable\n")
	sb.WriteString(fmt.Sprintf("export function %s<TType>(\n", unmaskFunctionName))
	sb.WriteString("  _documentNode: DocumentTypeDecoration<TType, any>,\n")
	sb.WriteString("  fragmentType: FragmentType<DocumentTypeDecoration<TType, any>>\n")
	sb.WriteString("): TType;\n")

	// Undefined overload
	sb.WriteString("// return nullable if `fragmentType` is undefined\n")
	sb.WriteString(fmt.Sprintf("export function %s<TType>(\n", unmaskFunctionName))
	sb.WriteString("  _documentNode: DocumentTypeDecoration<TType, any>,\n")
	sb.WriteString("  fragmentType: FragmentType<DocumentTypeDecoration<TType, any>> | undefined\n")
	sb.WriteString("): TType | undefined;\n")

	// Null overload
	sb.WriteString("// return nullable if `fragmentType` is nullable\n")
	sb.WriteString(fmt.Sprintf("export function %s<TType>(\n", unmaskFunctionName))
	sb.WriteString("  _documentNode: DocumentTypeDecoration<TType, any>,\n")
	sb.WriteString("  fragmentType: FragmentType<DocumentTypeDecoration<TType, any>> | null\n")
	sb.WriteString("): TType | null;\n")

	// Null or undefined overload
	sb.WriteString("// return nullable if `fragmentType` is nullable or undefined\n")
	sb.WriteString(fmt.Sprintf("export function %s<TType>(\n", unmaskFunctionName))
	sb.WriteString("  _documentNode: DocumentTypeDecoration<TType, any>,\n")
	sb.WriteString("  fragmentType: FragmentType<DocumentTypeDecoration<TType, any>> | null | undefined\n")
	sb.WriteString("): TType | null | undefined;\n")

	// Array overload
	sb.WriteString("// return array of non-nullable if `fragmentType` is array of non-nullable\n")
	sb.WriteString(fmt.Sprintf("export function %s<TType>(\n", unmaskFunctionName))
	sb.WriteString("  _documentNode: DocumentTypeDecoration<TType, any>,\n")
	sb.WriteString("  fragmentType: Array<FragmentType<DocumentTypeDecoration<TType, any>>>\n")
	sb.WriteString("): Array<TType>;\n")

	// Nullable array overload
	sb.WriteString("// return array of nullable if `fragmentType` is array of nullable\n")
	sb.WriteString(fmt.Sprintf("export function %s<TType>(\n", unmaskFunctionName))
	sb.WriteString("  _documentNode: DocumentTypeDecoration<TType, any>,\n")
	sb.WriteString("  fragmentType: Array<FragmentType<DocumentTypeDecoration<TType, any>>> | null | undefined\n")
	sb.WriteString("): Array<TType> | null | undefined;\n")

	// ReadonlyArray overload
	sb.WriteString("// return readonly array of non-nullable if `fragmentType` is array of non-nullable\n")
	sb.WriteString(fmt.Sprintf("export function %s<TType>(\n", unmaskFunctionName))
	sb.WriteString("  _documentNode: DocumentTypeDecoration<TType, any>,\n")
	sb.WriteString("  fragmentType: ReadonlyArray<FragmentType<DocumentTypeDecoration<TType, any>>>\n")
	sb.WriteString("): ReadonlyArray<TType>;\n")

	// Nullable ReadonlyArray overload
	sb.WriteString("// return readonly array of nullable if `fragmentType` is array of nullable\n")
	sb.WriteString(fmt.Sprintf("export function %s<TType>(\n", unmaskFunctionName))
	sb.WriteString("  _documentNode: DocumentTypeDecoration<TType, any>,\n")
	sb.WriteString("  fragmentType: ReadonlyArray<FragmentType<DocumentTypeDecoration<TType, any>>> | null | undefined\n")
	sb.WriteString("): ReadonlyArray<TType> | null | undefined;")
}

// writeUnmaskFunction writes the complete unmask function with implementation
func (p *Plugin) writeUnmaskFunction(sb *strings.Builder, unmaskFunctionName string) {
	// Write type definitions first
	p.writeUnmaskFunctionTypeDefinitions(sb, unmaskFunctionName)

	sb.WriteString("\n")

	// Implementation
	sb.WriteString(fmt.Sprintf("export function %s<TType>(\n", unmaskFunctionName))
	sb.WriteString("  _documentNode: DocumentTypeDecoration<TType, any>,\n")
	sb.WriteString("  fragmentType: FragmentType<DocumentTypeDecoration<TType, any>> | Array<FragmentType<DocumentTypeDecoration<TType, any>>> | ReadonlyArray<FragmentType<DocumentTypeDecoration<TType, any>>> | null | undefined\n")
	sb.WriteString("): TType | Array<TType> | ReadonlyArray<TType> | null | undefined {\n")
	sb.WriteString("  return fragmentType as any;\n")
	sb.WriteString("}")
}

// writeMakeFragmentDataHelper writes the makeFragmentData helper function
func (p *Plugin) writeMakeFragmentDataHelper(sb *strings.Builder) {
	sb.WriteString("export function makeFragmentData<\n")
	sb.WriteString("  F extends DocumentTypeDecoration<any, any>,\n")
	sb.WriteString("  FT extends ResultOf<F>\n")
	sb.WriteString(">(data: FT, _fragment: F): FragmentType<F> {\n")
	sb.WriteString("  return data as FragmentType<F>;\n")
	sb.WriteString("}")
}

// writeIsFragmentReadyFunction writes the isFragmentReady helper function
func (p *Plugin) writeIsFragmentReadyFunction(sb *strings.Builder, isStringDocumentMode bool) {
	if isStringDocumentMode {
		// String document mode version
		sb.WriteString("export function isFragmentReady<TQuery, TFrag>(\n")
		sb.WriteString("  queryNode: TypedDocumentString<TQuery, any>,\n")
		sb.WriteString("  fragmentNode: TypedDocumentString<TFrag, any>,\n")
		sb.WriteString("  data: FragmentType<TypedDocumentString<Incremental<TFrag>, any>> | null | undefined\n")
		sb.WriteString("): data is FragmentType<typeof fragmentNode> {\n")
		sb.WriteString("  const deferredFields = queryNode.__meta__?.deferredFields as Record<string, (keyof TFrag)[]>;\n")
		sb.WriteString("  const fragName = fragmentNode.__meta__?.fragmentName as string | undefined;\n")
		sb.WriteString("\n")
		sb.WriteString("  if (!deferredFields || !fragName) return true;\n")
		sb.WriteString("\n")
		sb.WriteString("  const fields = deferredFields[fragName] ?? [];\n")
		sb.WriteString("  return fields.length > 0 && fields.every(field => data && field in data);\n")
		sb.WriteString("}\n")
	} else {
		// Standard document mode version
		sb.WriteString("export function isFragmentReady<TQuery, TFrag>(\n")
		sb.WriteString("  queryNode: DocumentTypeDecoration<TQuery, any>,\n")
		sb.WriteString("  fragmentNode: TypedDocumentNode<TFrag>,\n")
		sb.WriteString("  data: FragmentType<TypedDocumentNode<Incremental<TFrag>, any>> | null | undefined\n")
		sb.WriteString("): data is FragmentType<typeof fragmentNode> {\n")
		sb.WriteString("  const deferredFields = (queryNode as { __meta__?: { deferredFields: Record<string, (keyof TFrag)[]> } }).__meta__\n")
		sb.WriteString("    ?.deferredFields;\n")
		sb.WriteString("\n")
		sb.WriteString("  if (!deferredFields) return true;\n")
		sb.WriteString("\n")
		sb.WriteString("  const fragDef = fragmentNode.definitions[0] as FragmentDefinitionNode | undefined;\n")
		sb.WriteString("  const fragName = fragDef?.name?.value;\n")
		sb.WriteString("\n")
		sb.WriteString("  const fields = (fragName && deferredFields[fragName]) || [];\n")
		sb.WriteString("  return fields.length > 0 && fields.every(field => data && field in data);\n")
		sb.WriteString("}\n")
	}
}