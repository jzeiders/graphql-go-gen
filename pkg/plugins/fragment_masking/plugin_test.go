package fragment_masking

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestPlugin_Name(t *testing.T) {
	p := &Plugin{}
	assert.Equal(t, "fragment-masking", p.Name())
}

func TestPlugin_Generate(t *testing.T) {
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name: "schema.graphql",
		Input: `
			type Query {
				user: User
			}
			type User {
				id: ID!
				name: String!
			}
		`,
	})
	require.NoError(t, err)

	t.Run("generates fragment masking utilities with default config", func(t *testing.T) {
		p := &Plugin{}
		output, err := p.Generate(schema, nil, nil)
		require.NoError(t, err)

		outputStr := string(output)

		// Check imports
		assert.Contains(t, outputStr, "import { ResultOf, DocumentTypeDecoration, TypedDocumentNode }")
		assert.Contains(t, outputStr, "from '@graphql-typed-document-node/core'")

		// Check FragmentType type definition
		assert.Contains(t, outputStr, "export type FragmentType<TDocumentType extends DocumentTypeDecoration<any, any>>")

		// Check makeFragmentData function
		assert.Contains(t, outputStr, "export function makeFragmentData<")

		// Check default useFragment function
		assert.Contains(t, outputStr, "export function useFragment<TType>(")

		// Check multiple overloads
		assert.Contains(t, outputStr, "Array<FragmentType<DocumentTypeDecoration<TType, any>>>")
		assert.Contains(t, outputStr, "| null | undefined")

		// Check isFragmentReady function
		assert.Contains(t, outputStr, "export function isFragmentReady<TQuery, TFrag>(")
	})

	t.Run("uses custom unmask function name", func(t *testing.T) {
		config := map[string]interface{}{
			"unmaskFunctionName": "readFragment",
		}

		p := &Plugin{}
		output, err := p.Generate(schema, nil, config)
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "export function readFragment<TType>(")
		assert.NotContains(t, outputStr, "export function useFragment<TType>(")
	})

	t.Run("uses type imports when configured", func(t *testing.T) {
		config := map[string]interface{}{
			"useTypeImports": true,
		}

		p := &Plugin{}
		output, err := p.Generate(schema, nil, config)
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "import type { ResultOf")
		assert.Contains(t, outputStr, "import type { FragmentDefinitionNode")
		assert.Contains(t, outputStr, "import type { Incremental")
	})

	t.Run("generates string document mode implementation", func(t *testing.T) {
		config := map[string]interface{}{
			"isStringDocumentMode": true,
		}

		p := &Plugin{}
		output, err := p.Generate(schema, nil, config)
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "const deferredFields = queryNode.definitions")
		assert.Contains(t, outputStr, "directive.name.value === 'defer'")
	})

	t.Run("generates document mode implementation", func(t *testing.T) {
		config := map[string]interface{}{
			"isStringDocumentMode": false,
		}

		p := &Plugin{}
		output, err := p.Generate(schema, nil, config)
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "__meta__?.deferredFields")
		assert.Contains(t, outputStr, "fields.every(field => data && field in data)")
	})
}

func TestPlugin_parseConfig(t *testing.T) {
	p := &Plugin{}

	t.Run("returns default config for nil", func(t *testing.T) {
		config := p.parseConfig(nil)
		assert.NotNil(t, config)
		assert.Equal(t, "", config.UnmaskFunctionName)
		assert.False(t, config.UseTypeImports)
		assert.False(t, config.EmitLegacyCommonJSImports)
		assert.False(t, config.IsStringDocumentMode)
	})

	t.Run("parses all config options", func(t *testing.T) {
		cfg := map[string]interface{}{
			"unmaskFunctionName":        "customUnmask",
			"useTypeImports":            true,
			"emitLegacyCommonJSImports": true,
			"isStringDocumentMode":      true,
		}

		config := p.parseConfig(cfg)
		assert.Equal(t, "customUnmask", config.UnmaskFunctionName)
		assert.True(t, config.UseTypeImports)
		assert.True(t, config.EmitLegacyCommonJSImports)
		assert.True(t, config.IsStringDocumentMode)
	})
}

func TestPlugin_GeneratedFunctions(t *testing.T) {
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name: "schema.graphql",
		Input: `type Query { id: ID }`,
	})
	require.NoError(t, err)

	p := &Plugin{}
	output, err := p.Generate(schema, nil, nil)
	require.NoError(t, err)

	outputStr := string(output)

	t.Run("has all useFragment overloads", func(t *testing.T) {
		// Count the number of useFragment function declarations
		count := strings.Count(outputStr, "export function useFragment<TType>(")
		assert.GreaterOrEqual(t, count, 4) // Should have at least 4 overloads + implementation
	})

	t.Run("has makeFragmentData function", func(t *testing.T) {
		assert.Contains(t, outputStr, "export function makeFragmentData<")
		assert.Contains(t, outputStr, "return data as FragmentType<F>;")
	})

	t.Run("has isFragmentReady function", func(t *testing.T) {
		assert.Contains(t, outputStr, "export function isFragmentReady<TQuery, TFrag>(")
		assert.Contains(t, outputStr, "data is FragmentType<DocumentTypeDecoration<TQuery, any>>")
	})
}

func TestPlugin_OutputValidity(t *testing.T) {
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name: "schema.graphql",
		Input: `type Query { id: ID }`,
	})
	require.NoError(t, err)

	configs := []map[string]interface{}{
		nil, // default config
		{"unmaskFunctionName": "readFragment"},
		{"useTypeImports": true},
		{"isStringDocumentMode": true},
		{
			"unmaskFunctionName":   "getFragment",
			"useTypeImports":       true,
			"isStringDocumentMode": true,
		},
	}

	for i, config := range configs {
		t.Run(fmt.Sprintf("config_%d", i), func(t *testing.T) {
			p := &Plugin{}
			output, err := p.Generate(schema, nil, config)
			require.NoError(t, err)

			outputStr := string(output)

			// Basic validity checks
			assert.True(t, strings.HasPrefix(outputStr, "/* eslint-disable */"))
			assert.Contains(t, outputStr, "export type FragmentType")
			assert.Contains(t, outputStr, "export function")
			assert.Contains(t, outputStr, "return")

			// Check for proper TypeScript syntax
			assert.NotContains(t, outputStr, "undefined undefined")
			assert.NotContains(t, outputStr, "null null")
		})
	}
}